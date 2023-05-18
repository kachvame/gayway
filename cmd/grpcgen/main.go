package main

import (
	"fmt"
	"github.com/kachvame/gayway/codegen"
	gaywayLog "github.com/kachvame/gayway/log"
	"github.com/rs/zerolog/log"
	"os"
)

func run() error {
	logLevel := os.Getenv("GAYWAY_LOG_LEVEL")
	dev := os.Getenv("GAYWAY_DEV")

	if logLevel == "" {
		logLevel = "info"
	}

	if err := gaywayLog.SetupLogger(logLevel, dev == "true"); err != nil {
		return fmt.Errorf("failed to set up logging: %w", err)
	}

	generatorLogger := log.With().Str("component", "grpcgen").Logger()

	generator := codegen.NewGenerator(&codegen.Config{
		Logger: generatorLogger,
	})

	if err := generator.Run(); err != nil {
		return fmt.Errorf("failed to run generator: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
}
