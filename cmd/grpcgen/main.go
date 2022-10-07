package main

import (
	_ "embed"
	"fmt"
	"github.com/kachvame/gayway/grpcgen"
	gaywayLog "github.com/kachvame/gayway/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
)

const (
	packageName      = "github.com/bwmarrin/discordgo"
	entrypointStruct = "Session"
)

var (
	ignoredMethods = map[string]struct{}{
		"AddHandler":     {},
		"AddHandlerOnce": {},
		"Open":           {},
		"Close":          {},
		"CloseWithCode":  {},
	}
)

func run() error {
	logLevel := os.Getenv("GAYWAY_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	if err := gaywayLog.SetupLogger(logLevel); err != nil {
		return fmt.Errorf("failed to set up logging: %w", err)
	}

	grpcgenLogger := log.With().Str("component", "grpcgen").Logger()

	generator := grpcgen.NewGenerator(grpcgenLogger, packageName, entrypointStruct, ignoredMethods)

	if err := generator.Run(); err != nil {
		return fmt.Errorf("generator failed: %w", err)
	}

	return nil
}

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DefaultContextLogger = &log.Logger

	if err := run(); err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
}
