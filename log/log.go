package log

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func SetupLogger(logLevel string) error {
	parsedLogLevel, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level '%s': %w", logLevel, err)
	}
	log.Logger = log.Level(parsedLogLevel)

	if os.Getenv("GAYWAY_DEV") == "true" {
		log.Logger = log.
			Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFormatUnix}).
			With().
			Caller().
			Logger()
	}

	return nil
}
