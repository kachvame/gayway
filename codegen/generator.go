package codegen

import "github.com/rs/zerolog"

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

	return nil
}
