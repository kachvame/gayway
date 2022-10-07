package grpcgen

import (
	_ "embed"
	"fmt"
	"github.com/kachvame/gayway/codegen"
	"github.com/rs/zerolog"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
)

type Generator struct {
	logger           zerolog.Logger
	packageName      string
	entrypointStruct string
	typesScope       *types.Scope
	ignoredMethods   map[string]struct{}
}

var (
	//go:embed templates/gayway.proto.tmpl
	protobufSchemaTemplateData string
)

func NewGenerator(logger zerolog.Logger, packageName, entrypointStruct string, ignoredMethods map[string]struct{}) *Generator {
	return &Generator{
		logger:           logger,
		packageName:      packageName,
		entrypointStruct: entrypointStruct,
		ignoredMethods:   ignoredMethods,
	}
}

func (generator *Generator) Run() error {
	packageConfig := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}

	discordgoPackages, err := packages.Load(packageConfig, generator.packageName)
	if err != nil {
		return fmt.Errorf("failed to load discordgo package: %w", err)
	}

	if packages.PrintErrors(discordgoPackages) > 0 {
		return fmt.Errorf("errors while loading discordgo package")
	}

	generator.logger.Info().Msg("parsed discordgo package")

	discordgoPackage := discordgoPackages[0]
	generator.typesScope = discordgoPackage.Types.Scope()

	_, sessionStruct := codegen.LookupType[*types.Named](generator.typesScope, generator.entrypointStruct)
	if sessionStruct == nil {
		return fmt.Errorf("failed to find entrypoint struct '%s'", generator.entrypointStruct)
	}

	generator.logger.Info().Msgf("found entrypoint struct '%s'", generator.entrypointStruct)

	for methodIdx := 0; methodIdx < sessionStruct.NumMethods(); methodIdx++ {
		method := sessionStruct.Method(methodIdx)

		if !method.Exported() {
			continue
		}

		methodName := method.Name()

		if _, ok := generator.ignoredMethods[methodName]; ok {
			continue
		}

		methodSignature := method.Type().(*types.Signature)
		methodParams := methodSignature.Params()
		methodResults := methodSignature.Results()

		generator.logger.Trace().
			Str("params", methodParams.String()).
			Str("results", methodResults.String()).
			Msgf("processing method %d: %s", methodIdx, methodName)

		generator.processTuple(methodParams)
		generator.processTuple(methodResults)
	}

	return nil
}

func (generator *Generator) processTuple(tuple *types.Tuple) {
 // TODO: implement
}
