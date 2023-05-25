package codegen

import (
	"fmt"
	"github.com/rs/zerolog"
	"go/types"
	"os"
	"os/exec"
)

var (
	ignoredMethods = map[string]struct{}{
		"AddHandler":              {},
		"AddHandlerOnce":          {},
		"Open":                    {},
		"Close":                   {},
		"CloseWithCode":           {},
		"ChannelVoiceJoin":        {},
		"RequestWithLockedBucket": {},
	}
)

const (
	outputFile    = "gateway.proto"
	outputPackage = "./grpc"
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

	protobufBuilder := NewProtobufBuilder("Gayway")
	protobufBuilder.SetOption("go_package", outputPackage)

	err = VisitExportedMethods(sessionStruct, func(method *types.Func) error {
		methodName := method.Name()
		if _, isIgnored := ignoredMethods[methodName]; isIgnored {
			generator.logger.Debug().Str("method", methodName).Msg("ignoring method")

			return nil
		}

		generator.logger.Debug().Str("method", methodName).Msg("visiting method")

		methodSignature := method.Type().(*types.Signature)

		paramTypes := tupleElements(methodSignature.Params())
		// NOTE: discordgo has a middleware-type variadic argument for all
		//       methods. We can't serialize them, so we need to ignore them.
		if methodSignature.Variadic() {
			paramTypes = paramTypes[:len(paramTypes)-1]
		}

		params, err := MessageFromFields(fmt.Sprintf("%sInput", methodName), paramTypes)
		if err != nil {
			return fmt.Errorf("failed to build protobuf params type: %w", err)
		}

		result, err := MessageFromFields(fmt.Sprintf("%sOutput", methodName), tupleElements(methodSignature.Results()))
		if err != nil {
			return fmt.Errorf("failed to build protobuf result type: %w", err)
		}

		protobufBuilder.AddMethod(methodName, params, result)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to build protobuf schema: %w", err)
	}

	proto := protobufBuilder.Serialize()
	if err = os.WriteFile(outputFile, []byte(proto), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	generator.logger.Info().Msgf("Wrote output to %s", outputFile)

	if err = exec.Command("protoc", "--go-grpc_out=.", "--go_out=.", outputFile).Run(); err != nil {
		return fmt.Errorf("failed to generate grpc interfaces: %w", err)
	}

	generator.logger.Info().Msgf("Generated gRPC boilerplate in %s", outputPackage)

	return nil
}

func tupleElements(tuple *types.Tuple) []*types.Var {
	elements := make([]*types.Var, 0, tuple.Len())
	_ = VisitTupleElements(tuple, func(_ int, variable *types.Var) error {
		elements = append(elements, variable)

		return nil
	})

	return elements
}
