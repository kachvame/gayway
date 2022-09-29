package main

import (
	"fmt"
	"github.com/kachvame/gayway/codegen"
	gaywayLog "github.com/kachvame/gayway/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
)

const (
	packageName = "github.com/bwmarrin/discordgo"
)

func run() error {
	logLevel := os.Getenv("GAYWAY_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	if err := gaywayLog.SetupLogger(logLevel); err != nil {
		return fmt.Errorf("failed to set up logging: %w", err)
	}

	packageConfig := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}

	discordgoPackages, err := packages.Load(packageConfig, packageName)
	if err != nil {
		return fmt.Errorf("failed to load discordgo package: %w", err)
	}

	if packages.PrintErrors(discordgoPackages) > 0 {
		return fmt.Errorf("errors while loading discordgo package")
	}

	log.Info().Msg("parsed discordgo package")

	discordgoPackage := discordgoPackages[0]

	structs := codegen.FindStructs(discordgoPackage)

	sessionStruct := structs["Session"]
	if sessionStruct == nil {
		return fmt.Errorf("failed to find 'Session' struct in discordgo package")
	}

	log.Info().Msg("found 'Session' struct")

	for _, method := range sessionStruct.Methods {
		log.Info().Msgf("found method '%s'", method.Name)
	}

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
