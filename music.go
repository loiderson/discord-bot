package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/disgolink/v4/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

var lava *disgolink.Client

// ---------- The queue: one track list per server, mutex-guarded ----------
// This is the Store pattern from Part 6: track-end events arrive on
// different goroutines than commands, so shared state needs the lock.

var (
	queueMu sync.Mutex
	queues  = map[string][]lavalink.Track{}
)

// queueAdd appends a track and returns its position in line.
func queueAdd(guildID string, t lavalink.Track) int {
	queueMu.Lock()
	defer queueMu.Unlock()
	queues[guildID] = append(queues[guildID], t)
	return len(queues[guildID])
}

// queueNext pops the front track. Comma-ok shape: (track, was there one?).
func queueNext(guildID string) (lavalink.Track, bool) {
	queueMu.Lock()
	defer queueMu.Unlock()
	q := queues[guildID]
	if len(q) == 0 {
		return lavalink.Track{}, false
	}
	next := q[0]
	queues[guildID] = q[1:]
	return next, true
}

func queueClear(guildID string) {
	queueMu.Lock()
	defer queueMu.Unlock()
	delete(queues, guildID)
}

// queueList returns a copy, so callers can read it without holding the lock.
func queueList(guildID string) []lavalink.Track {
	queueMu.Lock()
	defer queueMu.Unlock()
	q := queues[guildID]
	out := make([]lavalink.Track, len(q))
	copy(out, q)
	return out
}

// ---------- Lavalink setup ----------

func initLavalink(s *discordgo.Session) error {
	lava = disgolink.New(snowflake.MustParse(s.State.User.ID),
		disgolink.WithListenerFunc(onTrackEnd),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := lava.AddNode(ctx, disgolink.NodeConfig{
		Name:     "local",
		Address:  "localhost:2333",
		Password: "REDACTED",
		Secure:   false,
	})
	if err != nil {
		return fmt.Errorf("could not connect to Lavalink (is it running?): %w", err)
	}
	return nil
}

// onTrackEnd fires when a track finishes for ANY reason.
// If it ended naturally, advance to the next queued track.
func onTrackEnd(event *disgolink.PlayerTrackEndEvent) {
	if !event.Reason.MayStartNext() {
		return // stopped/replaced on purpose — don't auto-advance
	}
	next, ok := queueNext(event.GetGuildID().String())
	if !ok {
		return // queue empty; playback simply ends
	}
	if err := event.Player.Update(context.TODO(), disgolink.WithTrack(next)); err != nil {
		fmt.Println("failed to play next track:", err)
	}
}

// ---------- Voice event forwarding (unchanged) ----------

func onVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if lava == nil || e.UserID != s.State.User.ID {
		return
	}
	var channelID *snowflake.ID
	if e.ChannelID != "" {
		id := snowflake.MustParse(e.ChannelID)
		channelID = &id
	}
	lava.OnVoiceStateUpdate(context.TODO(), snowflake.MustParse(e.GuildID), channelID, e.SessionID)
}

func onVoiceServerUpdate(s *discordgo.Session, e *discordgo.VoiceServerUpdate) {
	if lava == nil {
		return
	}
	lava.OnVoiceServerUpdate(context.TODO(), snowflake.MustParse(e.GuildID), e.Token, e.Endpoint)
}

// ---------- Commands ----------

func cmdPlay(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "usage: !play <youtube link or search words>")
		return
	}

	vs, err := s.State.VoiceState(m.GuildID, m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "you need to be in a voice channel first")
		return
	}

	query := strings.Join(args, " ")
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
		query = lavalink.SearchTypeYouTube.Apply(query)
	}

	player := lava.Player(snowflake.MustParse(m.GuildID))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var toPlay *lavalink.Track
	lava.BestNode().Rest.LoadTracksHandler(ctx, query, disgolink.NewTrackLoadingResultHandler(
		func(track lavalink.Track) {
			toPlay = &track
		},
		func(playlist lavalink.Playlist) {
			toPlay = &playlist.Tracks[0]
		},
		func(tracks []lavalink.Track) {
			toPlay = &tracks[0]
		},
		func() {
			s.ChannelMessageSend(m.ChannelID, "nothing found for: "+strings.Join(args, " "))
		},
		func(err error) {
			s.ChannelMessageSend(m.ChannelID, "error looking up track: "+err.Error())
		},
	))
	if toPlay == nil {
		return
	}

	// Something already playing? Queue it instead of replacing it.
	if player.Track != nil {
		pos := queueAdd(m.GuildID, *toPlay)
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("➕ Queued **%s** (position %d)", toPlay.Info.Title, pos))
		return
	}

	if err := s.ChannelVoiceJoinManual(m.GuildID, vs.ChannelID, false, true); err != nil {
		s.ChannelMessageSend(m.ChannelID, "couldn't join voice: "+err.Error())
		return
	}

	if err := player.Update(context.TODO(), disgolink.WithTrack(*toPlay)); err != nil {
		s.ChannelMessageSend(m.ChannelID, "couldn't start playback: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "▶️ Now playing: "+toPlay.Info.Title)
}

func cmdSkip(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	player := lava.ExistingPlayer(snowflake.MustParse(m.GuildID))
	if player == nil || player.Track == nil {
		s.ChannelMessageSend(m.ChannelID, "nothing is playing")
		return
	}

	next, ok := queueNext(m.GuildID)
	if !ok {
		// nothing queued — just stop the current track (bot stays in channel)
		if err := player.Update(context.TODO(), disgolink.WithNullTrack()); err != nil {
			s.ChannelMessageSend(m.ChannelID, "error: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "⏭️ skipped — queue is empty")
		return
	}

	if err := player.Update(context.TODO(), disgolink.WithTrack(next)); err != nil {
		s.ChannelMessageSend(m.ChannelID, "error: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "⏭️ Now playing: "+next.Info.Title)
}

func cmdQueue(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	tracks := queueList(m.GuildID)
	if len(tracks) == 0 {
		s.ChannelMessageSend(m.ChannelID, "the queue is empty")
		return
	}
	var b strings.Builder
	b.WriteString("🎶 **Queue:**\n")
	for i, t := range tracks {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, t.Info.Title))
	}
	s.ChannelMessageSend(m.ChannelID, b.String())
}

func cmdPause(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	player := lava.ExistingPlayer(snowflake.MustParse(m.GuildID))
	if player == nil {
		s.ChannelMessageSend(m.ChannelID, "nothing is playing")
		return
	}
	if err := player.Update(context.TODO(), disgolink.WithPaused(!player.Paused)); err != nil {
		s.ChannelMessageSend(m.ChannelID, "error: "+err.Error())
		return
	}
	if player.Paused {
		s.ChannelMessageSend(m.ChannelID, "⏸️ paused — !pause again to resume")
	} else {
		s.ChannelMessageSend(m.ChannelID, "▶️ resumed")
	}
}

func cmdStop(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	player := lava.ExistingPlayer(snowflake.MustParse(m.GuildID))
	if player == nil {
		s.ChannelMessageSend(m.ChannelID, "nothing is playing")
		return
	}
	queueClear(m.GuildID) // stop means stop — drop everything queued too
	if err := s.ChannelVoiceJoinManual(m.GuildID, "", false, false); err != nil {
		s.ChannelMessageSend(m.ChannelID, "error while disconnecting: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "⏹️ stopped and cleared the queue")
}
