package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func cmdPing(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, codeBox("Pong!"))
}

func cmdEcho(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, codeBox("usage: !echo <text>"))
		return
	}
	s.ChannelMessageSend(m.ChannelID, codeBox(strings.Join(args, " ")))
}

func cmdRoll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	n := rand.Intn(6) + 1
	s.ChannelMessageSend(m.ChannelID, codeBox(fmt.Sprintf("🎲 You rolled a %d", n)))
}

func cmdHelp(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	help := `🎵  MUSIC BOT — COMMANDS

  Music
    !play <link or search>   play a track, or queue it if one is playing
    !pause                   pause the current track
    !unpause                 resume playback
    !skip                    skip to the next track in the queue
    !stop                    stop playback and leave the voice channel

  Queue
    !queue                   show the current queue
    !shuffle                 randomize the queue order
    !clear                   empty the queue (keeps current track playing)

  Misc
    !ping                    check the bot is alive
    !echo <text>             repeat your text back
    !roll                    roll a six-sided die
    !help                    show this message`

	s.ChannelMessageSend(m.ChannelID, codeBox(help))
}
