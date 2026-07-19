# Discord Music Bot (Go + Lavalink)

A Discord music bot written in Go using discordgo, with audio handled by a
local Lavalink server (required for DAVE/E2EE voice support since March 2026).

Plays YouTube links, YouTube searches, and Spotify track links
(resolved to YouTube audio via the LavaSrc plugin).

## Commands

`!play <link or search words>` — play or queue a track (joins your voice channel)
`!skip` / `!pause` / `!stop` / `!queue` — the usual suspects
`!ping`, `!echo`, `!roll` — leftovers from learning Go, kept out of affection

## Setup

### 1. Lavalink (the audio server)

Requires Java 17+.

    mkdir ../lavalink && cd ../lavalink
    wget https://github.com/lavalink-devs/Lavalink/releases/download/4.2.2/Lavalink.jar

Copy `application.example.yml` from this repo to `../lavalink/application.yml`
and fill in:
- a Lavalink password of your choosing
- Spotify Client ID + Secret (free app at developer.spotify.com/dashboard)

Run it (leave this terminal open):

    java -jar Lavalink.jar

### 2. The bot

- Create a Discord application (Guild Install), enable the Message Content
  intent, invite it with Send Messages / Read Message History / Connect / Speak.
- Put the bot token in `../bot-token.txt` (outside the repo on purpose),
  or point the TOKEN_FILE env var at it.
- Make sure the Lavalink password in `music.go`'s NodeConfig matches your yml.

    go run .

Bot in one terminal, Lavalink in the other. Both must be running.
