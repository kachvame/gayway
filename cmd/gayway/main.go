package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	gaywayLog "github.com/kachvame/gayway/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"os/signal"
	"syscall"
)

func run() error {
	logLevel := os.Getenv("GAYWAY_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	if err := gaywayLog.SetupLogger(logLevel); err != nil {
		return fmt.Errorf("failed to set up logging: %w", err)
	}

	discordLogger := log.With().Str("component", "bot").Logger()

	discord, err := discordgo.New("Bot " + os.Getenv("GAYWAY_DISCORD_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to initialize discordgo: %w", err)
	}

	if err := discord.Open(); err != nil {
		return fmt.Errorf("failed to open discord client: %w", err)
	}

	discordLogger.Info().Msg("opened discord client")

	defer func(discord *discordgo.Session) {
		_ = discord.Close()
	}(discord)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return nil
}

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DefaultContextLogger = &log.Logger

	if err := run(); err != nil {
		log.Error().Err(err)
		os.Exit(1)
	}
}
