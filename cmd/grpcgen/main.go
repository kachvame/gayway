package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/kachvame/gayway/codegen"
	gaywayLog "github.com/kachvame/gayway/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

const (
	packageName = "github.com/bwmarrin/discordgo"
)

var (
	//go:embed templates/gayway.proto.tmpl
	protobufSchemaTemplateData string

	blacklistedMethods = map[string]struct{}{
		"Open":           {},
		"Close":          {},
		"AddHandler":     {},
		"AddHandlerOnce": {},
		// FIXME: these methods use enums that aren't actually enums and those
		//        are impossible for us to detect statically, so unless the
		//        issue is fixed upstream, these should stay ignored.
		"GuildScheduledEvent":       {},
		"GuildScheduledEventCreate": {},
		"GuildScheduledEventDelete": {},
		"GuildScheduledEventEdit":   {},
		"GuildScheduledEventUsers":  {},
		"GuildScheduledEvents":      {},
	}

	fieldlessStructs = map[string]struct{}{
		"Bucket":          {},
		"State":           {},
		"RateLimiter":     {},
		"Session":         {},
		"VoiceConnection": {},
		// FIXME: these structs use enums that aren't actually enums and those
		//        are impossible for us to detect statically, so unless the
		//        issue is fixed upstream, these should stay ignored.
		"GuildScheduledEvent":       {},
		"GuildScheduledEventParams": {},
	}
)

type ProtobufSchemaContext struct {
	Methods    []*Method
	Structs    []*Struct
	Enums      []*Enum
	Interfaces []*Interface
}

type Method struct {
	Name   string
	Input  *Message
	Output *Message
}

type Struct struct {
	*Message
	Name string
}

type Message struct {
	Fields []*MessageField
}

type MessageField struct {
	Name string
	Type *codegen.TypeExpression
}

type Enum struct {
	Name         string
	Values       []*EnumValue
	StringValues []string
}

type EnumValue struct {
	Name  string
	Value int
}

type Interface struct {
	Name            string
	Implementations []string
}

func codegenValuesToProtoMessage(values []*codegen.Value) *Message {
	fields := make([]*MessageField, 0, len(values))
	for _, value := range values {
		messageField := &MessageField{
			Name: value.Name,
			Type: value.Type,
		}

		fields = append(fields, messageField)
	}

	return &Message{
		Fields: fields,
	}
}

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

	structsMap := codegen.FindStructs(discordgoPackage, fieldlessStructs)
	enumsMap := codegen.FindEnums(discordgoPackage)
	interfacesMap := codegen.FindInterfaces(discordgoPackage, structsMap)

	sessionStruct := structsMap["Session"]
	if sessionStruct == nil {
		return fmt.Errorf("failed to find 'Session' struct in discordgo package")
	}

	log.Info().Msg("found 'Session' struct")

	methods := make([]*Method, 0, len(sessionStruct.Methods)-len(blacklistedMethods))
	for _, methodDeclaration := range sessionStruct.Methods {
		if _, isBlacklisted := blacklistedMethods[methodDeclaration.Name]; isBlacklisted {
			continue
		}

		log.Info().Msgf("processing method '%s'", methodDeclaration.Name)

		method := &Method{
			Name:   methodDeclaration.Name,
			Input:  codegenValuesToProtoMessage(methodDeclaration.Args),
			Output: codegenValuesToProtoMessage(methodDeclaration.Results),
		}

		methods = append(methods, method)
	}

	structs := make([]*Struct, 0, len(structsMap))
	for name, structDeclaration := range structsMap {
		structs = append(structs, &Struct{
			Name:    name,
			Message: codegenValuesToProtoMessage(structDeclaration.Fields),
		})
	}

	enums := make([]*Enum, 0, len(enumsMap))
	for name, enumValues := range enumsMap {
		values := make([]*EnumValue, 0, len(enumValues))

		var stringValues []string
		seen := make(map[int]struct{})

		hasZero := false
		for idx, enumValue := range enumValues {
			name := enumValue.Name()
			value := enumValue.Val().ExactString()

			numericValue, err := strconv.Atoi(value)
			if err != nil {
				if stringValues == nil {
					stringValues = make([]string, 0, len(enumValues))
				}

				numericValue = idx
				stringValues = append(stringValues, value)
			}

			if numericValue == 0 {
				hasZero = true
			}

			if _, ok := seen[numericValue]; ok {
				continue
			}

			seen[numericValue] = struct{}{}

			values = append(values, &EnumValue{
				Name:  name,
				Value: numericValue,
			})
		}

		if !hasZero {
			values = append([]*EnumValue{
				{
					Name:  fmt.Sprintf("%sInternal", name),
					Value: 0,
				},
			}, values...)
		}

		sort.Slice(values, func(i, j int) bool {
			return values[i].Value < values[j].Value
		})

		enums = append(enums, &Enum{
			Name:         name,
			Values:       values,
			StringValues: stringValues,
		})
	}

	interfaces := make([]*Interface, 0, len(interfacesMap))
	for name, implementations := range interfacesMap {
		interfaces = append(interfaces, &Interface{
			Name:            name,
			Implementations: implementations,
		})
	}

	protobufSchemaTemplate, err := template.New("protobufSchema").Funcs(template.FuncMap{
		"add": func(x, y int) int {
			return x + y
		},
		"mapType": func(typ *codegen.TypeExpression) string {
			result := &strings.Builder{}

			if typ.IsArray {
				result.WriteString("repeated ")
			}

			var name string
			if typ.Identifier == nil {
				name = "google.protobuf.Any"
			} else {
				name = typ.Identifier.Name
			}

			switch name {
			case "int", "int16", "byte":
				result.WriteString("int32")
			case "uint16":
				result.WriteString("uint32")
			case "float64":
				result.WriteString("double")
			case "error":
				result.WriteString("string")
			default:
				result.WriteString(name)
			}

			return result.String()
		},
	}).Parse(protobufSchemaTemplateData)
	if err != nil {
		return fmt.Errorf("failed to parse protobuf schema template: %w", err)
	}

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	sort.Slice(structs, func(i, j int) bool {
		return structs[i].Name < structs[j].Name
	})

	sort.Slice(enums, func(i, j int) bool {
		return enums[i].Name < enums[j].Name
	})

	protobufSchemaContext := &ProtobufSchemaContext{
		Methods:    methods,
		Structs:    structs,
		Enums:      enums,
		Interfaces: interfaces,
	}

	outputBuffer := &bytes.Buffer{}
	if err := protobufSchemaTemplate.Execute(outputBuffer, protobufSchemaContext); err != nil {
		return fmt.Errorf("failed to render protobuf schema template: %w", err)
	}

	fmt.Println(outputBuffer.String())

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
