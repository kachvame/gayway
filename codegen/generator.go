package codegen

import (
	"fmt"
	"github.com/rs/zerolog"
	"go/types"
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

type Config struct {
	Logger zerolog.Logger
}

type Generator struct {
	logger zerolog.Logger
}

func NewGenerator(config *Config) *Generator {
	return &Generator{
		logger: config.Logger,
	}
}

func (generator *Generator) Run() error {
	generator.logger.Info().Msg("Hello world")

	discordgoPackage, err := generator.loadDiscordgoPackage()
	if err != nil {
		return fmt.Errorf("failed to load discordgo package: %w", err)
	}

	generator.logger.Debug().Msg("loaded discordgo package")

	sessionStruct, err := generator.findSessionStruct(discordgoPackage)
	if err != nil {
		return fmt.Errorf("failed to find Session struct: %w", err)
	}

	generator.logger.Debug().Msg("found Session struct")

	VisitExportedMethods(sessionStruct, func(method *types.Func) {
		methodName := method.Name()
		if _, isIgnored := ignoredMethods[methodName]; isIgnored {
			generator.logger.Debug().Str("method", methodName).Msg("ignoring method")

			return
		}

		generator.logger.Debug().Str("method", methodName).Msg("visiting method")
	})

	return nil
}
