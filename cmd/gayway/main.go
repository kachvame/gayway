package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func run() error {
	discord, err := discordgo.New("Bot " + os.Getenv("GAYWAY_DISCORD_TOKEN"))
	if err != nil {
		return err
	}

	if err := discord.Open(); err != nil {
		return err
	}

	log.Println("opened discord client")

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
		log.Fatalln("gayway: ", err)
	}
}
