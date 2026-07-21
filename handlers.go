package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type commandFunc func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)

var commands = map[string]commandFunc{
	"ping":    cmdPing,
	"echo":    cmdEcho,
	"roll":    cmdRoll,
	"play":    cmdPlay,
	"stop":    cmdStop,
	"pause":   cmdPause,
	"unpause": cmdUnpause,
	"clear":   cmdClear,
	"skip":    cmdSkip,
	"queue":   cmdQueue,
	"help":    cmdHelp,
	"shuffle": cmdShuffle,
	"loop":    cmdLoop,
	"unloop":  cmdUnloop,
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	parts := strings.Fields(m.Content[1:]) // drop the "!", split on whitespace
	if len(parts) == 0 {
		return
	}

	name, args := parts[0], parts[1:]
	if fn, ok := commands[name]; ok {
		fn(s, m, args)
	}
}
