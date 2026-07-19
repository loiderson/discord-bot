package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func loadSecret(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read token file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func main() {
	path := os.Getenv("TOKEN_FILE")
	if path == "" {
		path = "../bot-token.txt"
	}

	token, err := loadSecret(path)
	if err != nil {
		fmt.Println("Error loading token:", err)
		return
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	session.State.TrackVoice = true
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentMessageContent

	session.AddHandler(messageHandler) // already there
	session.AddHandler(onVoiceStateUpdate)
	session.AddHandler(onVoiceServerUpdate)

	err = session.Open()
	if err != nil {
		fmt.Println("Error opening Discord session:", err)
		return
	}
	if err := initLavalink(session); err != nil {
		fmt.Println("Error connecting to Lavalink:", err)
		return
	}
	defer session.Close()

	fmt.Println("Bot is running. Press Ctrl+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("Bot is shutting down.")
}
