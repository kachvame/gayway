package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func run() error {
	discord, err := discordgo.New("Bot " + os.Getenv("GAYWAY_DISCORD_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to initialize discordgo: %w", err)
	}

	if err = discord.Open(); err != nil {
		return fmt.Errorf("failed to open discord client: %w", err)
	}

	defer func(discord *discordgo.Session) {
		_ = discord.Close()
	}(discord)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
