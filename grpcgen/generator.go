package grpcgen

import (
	_ "embed"
	"github.com/rs/zerolog"
)

type Generator struct {
	logger           zerolog.Logger
	packageName      string
	entrypointStruct string
}

var (
	//go:embed templates/gayway.proto.tmpl
	protobufSchemaTemplateData string
)

func NewGenerator(logger zerolog.Logger, packageName, entrypointStruct string) *Generator {
	return &Generator{
		logger:           logger,
		packageName:      packageName,
		entrypointStruct: entrypointStruct,
	}
}

func (generator *Generator) Run() error {
	return nil
}
