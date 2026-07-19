package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func cmdPing(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "Pong!")
}

func cmdEcho(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, strings.Join(args, " "))
}

func cmdRoll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	n := rand.Intn(6) + 1
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎲 You rolled a %d", n))
}

func cmdHelp(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "Available commands: ping, echo, roll, help, play, pause, skip, queue, stop")
}
