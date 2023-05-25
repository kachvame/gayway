package codegen

import (
	"fmt"
	"go/types"
	"sort"
	"strconv"
	"strings"
)

type ProtobufBuilder struct {
	Name    string
	Methods []*ProtobufMethod
	Options map[string]string
}

func NewProtobufBuilder(serviceName string) *ProtobufBuilder {
	return &ProtobufBuilder{
		Name:    serviceName,
		Options: make(map[string]string),
	}
}

func (builder *ProtobufBuilder) AddMethod(name string, argument ProtobufType, returnType ProtobufType) {
	builder.Methods = append(builder.Methods, &ProtobufMethod{
		Name:       name,
		Argument:   argument,
		ReturnType: returnType,
	})
}

func (builder *ProtobufBuilder) SetOption(key string, value string) {
	builder.Options[key] = value
}

func (builder *ProtobufBuilder) Serialize() string {
	outputBuilder := &strings.Builder{}

	outputBuilder.WriteString(`syntax = "proto3";`)
	outputBuilder.WriteRune('\n')
	// Always import `any`
	outputBuilder.WriteString(`import "google/protobuf/any.proto";`)
	outputBuilder.WriteRune('\n')

	// Serialize options
	for key, value := range builder.Options {
		outputBuilder.WriteString("option ")
		outputBuilder.WriteString(key)
		outputBuilder.WriteString(" = ")
		outputBuilder.WriteString(strconv.Quote(value))
		outputBuilder.WriteString(";\n")
	}

	// Serialize service
	outputBuilder.WriteString("service ")
	outputBuilder.WriteString(builder.Name)
	outputBuilder.WriteString(" {\n")

	for _, method := range builder.Methods {
		outputBuilder.WriteString("  rpc ")
		outputBuilder.WriteString(method.Name)

		outputBuilder.WriteRune('(')
		outputBuilder.WriteString(method.Argument.Reference())
		outputBuilder.WriteRune(')')

		outputBuilder.WriteString(" returns (")
		outputBuilder.WriteString(method.ReturnType.Reference())
		outputBuilder.WriteString(");\n")
	}
	outputBuilder.WriteString("}\n")

	// Serialize all types
	serializableTypeSet := make(map[ProtobufSerializable]struct{})

	var visitType func(targetType ProtobufType)
	visitType = func(targetType ProtobufType) {
		if serializable, ok := targetType.(ProtobufSerializable); ok {
			serializableTypeSet[serializable] = struct{}{}
		}

		if visitor, ok := targetType.(ProtobufVisitor); ok {
			visitor.Visit(visitType)
		}
	}

	for _, method := range builder.Methods {
		visitType(method.Argument)
		visitType(method.ReturnType)
	}

	serializableTypes := make([]ProtobufSerializable, 0, len(serializableTypeSet))
	for value := range serializableTypeSet {
		serializableTypes = append(serializableTypes, value)
	}

	sort.Slice(serializableTypes, func(i, j int) bool {
		return serializableTypes[i].Reference() < serializableTypes[j].Reference()
	})

	for _, value := range serializableTypes {
		value.Serialize(outputBuilder)
		outputBuilder.WriteRune('\n')
	}

	return outputBuilder.String()
}

type ProtobufMethod struct {
	Name       string
	Argument   ProtobufType
	ReturnType ProtobufType
}

type ProtobufType interface {
	Reference() string
}

type ProtobufSerializable interface {
	ProtobufType

	Serialize(builder *strings.Builder)
}

type ProtobufVisitor interface {
	Visit(visitor func(node ProtobufType))
}

type ProtobufString struct{}

func (str *ProtobufString) Reference() string {
	return "string"
}

type ProtobufUint32 struct{}

func (num *ProtobufUint32) Reference() string {
	return "uint32"
}

type ProtobufInt32 struct{}

func (num *ProtobufInt32) Reference() string {
	return "int32"
}

type ProtobufUint64 struct{}

func (num *ProtobufUint64) Reference() string {
	return "uint64"
}

type ProtobufInt64 struct{}

func (num *ProtobufInt64) Reference() string {
	return "int64"
}

type ProtobufFloat struct{}

func (num *ProtobufFloat) Reference() string {
	return "float"
}

type ProtobufDouble struct{}

func (num *ProtobufDouble) Reference() string {
	return "double"
}

type ProtobufBoolean struct{}

func (str *ProtobufBoolean) Reference() string {
	return "bool"
}

type ProtobufOptional struct {
	Element ProtobufType
}

func (optional *ProtobufOptional) Reference() string {
	return fmt.Sprintf("optional %s", optional.Element.Reference())
}

func (optional *ProtobufOptional) Visit(visitor func(node ProtobufType)) {
	visitor(optional.Element)
}

type ProtobufArray struct {
	Element ProtobufType
}

func (array *ProtobufArray) Reference() string {
	return fmt.Sprintf("repeated %s", array.Element.Reference())
}

func (array *ProtobufArray) Visit(visitor func(node ProtobufType)) {
	visitor(array.Element)
}

type ProtobufMap struct {
	KeyType   ProtobufType
	ValueType ProtobufType
}

func (mapType *ProtobufMap) Reference() string {
	return fmt.Sprintf("map<%s, %s>", mapType.KeyType.Reference(), mapType.ValueType.Reference())
}

func (mapType *ProtobufMap) Visit(visitor func(node ProtobufType)) {
	visitor(mapType.KeyType)
	visitor(mapType.ValueType)
}

type ProtobufMessage struct {
	Name   string
	Fields []*ProtobufMessageField
}

type ProtobufMessageField struct {
	Name string
	Type ProtobufType
}

func (message *ProtobufMessage) Reference() string {
	return message.Name
}

func (message *ProtobufMessage) Serialize(builder *strings.Builder) {
	builder.WriteString("message ")
	builder.WriteString(message.Name)
	builder.WriteString(" {\n")

	for idx, field := range message.Fields {
		builder.WriteString("  ")
		builder.WriteString(field.Type.Reference())
		builder.WriteRune(' ')
		builder.WriteString(field.Name)
		builder.WriteString(" = ")
		builder.WriteString(strconv.Itoa(idx + 1))
		builder.WriteString(";\n")
	}

	builder.WriteString("}")
}

func (message *ProtobufMessage) Visit(visitor func(node ProtobufType)) {
	for _, field := range message.Fields {
		visitor(field.Type)
	}
}

type ProtobufAny struct{}

func (any *ProtobufAny) Reference() string {
	return "google.protobuf.Any"
}

type ProtobufReference struct {
	Type types.Type
}

func (reference *ProtobufReference) Reference() string {
	actualType, err := NewProtobufType(reference.Type)
	if err != nil {
		panic(err)
	}

	return actualType.Reference()
}

func MessageFromFields(name string, fields []*types.Var) (*ProtobufMessage, error) {
	message := &ProtobufMessage{
		Name: name,
	}

	for idx, field := range fields {
		fieldName := field.Name()
		if fieldName == "" {
			fieldName = fmt.Sprintf("field%d", idx+1)
		}

		fieldType, err := NewProtobufType(field.Type())
		if err != nil {
			return nil, fmt.Errorf("failed to convert element %d of tuple %s: %w", idx, name, err)
		}

		message.Fields = append(message.Fields, &ProtobufMessageField{
			Name: fieldName,
			Type: fieldType,
		})
	}

	return message, nil
}

var (
	protobufTypeCache = make(map[types.Type]ProtobufType)
)

func NewProtobufType(targetType types.Type) (ProtobufType, error) {
	if existingType, ok := protobufTypeCache[targetType]; ok {
		// NOTE: We've encountered a circular reference. To handle this,
		//       return a "reference" type which should be resolvable later.
		if existingType == nil {
			return &ProtobufReference{
				Type: targetType,
			}, nil
		}

		return existingType, nil
	}

	protobufTypeCache[targetType] = nil

	result, err := newProtobufType(targetType)
	if err != nil {
		delete(protobufTypeCache, targetType)

		return nil, err
	}

	protobufTypeCache[targetType] = result
	return result, nil
}

func newProtobufType(targetType types.Type) (ProtobufType, error) {
	switch targetType.(type) {
	case *types.Pointer:
		pointer := targetType.(*types.Pointer)

		return NewProtobufType(pointer.Elem())
	case *types.Named:
		named := targetType.(*types.Named)

		return convertNamed(named)
	case *types.Basic:
		basicType := targetType.(*types.Basic)

		return convertBasicType(basicType)
	// NOTE: we treat getting an unnamed interface as any
	case *types.Interface:
		return &ProtobufAny{}, nil
	case *types.Slice:
		slice := targetType.(*types.Slice)

		return convertSlice(slice)
	case *types.Map:
		mapType := targetType.(*types.Map)

		return convertMap(mapType)
	}

	return nil, fmt.Errorf("can't create protobuf type: %v (%T)", targetType, targetType)
}

func convertBasicType(basicType *types.Basic) (ProtobufType, error) {
	switch basicType.Kind() {
	case types.String:
		return &ProtobufString{}, nil
	case types.Bool:
		return &ProtobufBoolean{}, nil
	case types.Int:
		fallthrough
	case types.Int32:
		return &ProtobufInt32{}, nil
	case types.Uint:
		fallthrough
	// NOTE: bytes are encoded as full uint32s as Protobuf has no uint8
	case types.Byte:
		fallthrough
	case types.Uint32:
		return &ProtobufUint32{}, nil
	case types.Int64:
		return &ProtobufInt64{}, nil
	case types.Uint64:
		return &ProtobufUint64{}, nil
	case types.Float32:
		return &ProtobufFloat{}, nil
	case types.Float64:
		return &ProtobufDouble{}, nil
	}

	return nil, fmt.Errorf("unknown basic type: %v", basicType)
}

func IsErr(target types.Type) bool {
	named, ok := target.(*types.Named)
	if !ok {
		return false
	}

	if _, ok = named.Underlying().(*types.Interface); !ok {
		return false
	}

	isBuiltin := named.Obj().Pkg() == nil
	if !isBuiltin {
		return false
	}

	return named.Obj().Name() == "error"
}

func convertNamed(named *types.Named) (ProtobufType, error) {
	name := named.Obj().Name()
	underlying := named.Underlying()

	if IsErr(named) {
		return &ProtobufOptional{
			Element: &ProtobufString{},
		}, nil
	}

	switch underlying.(type) {
	case *types.Struct:
		return convertStruct(name, underlying.(*types.Struct))
	// TODO: interfaces
	case *types.Interface:
		return &ProtobufMessage{
			Name: name,
		}, nil
	// TODO: enums
	case *types.Basic:
		return convertBasicType(underlying.(*types.Basic))
	}

	return nil, fmt.Errorf("unknown named type: %v (%T)", named, underlying)
}

func convertStruct(name string, structType *types.Struct) (*ProtobufMessage, error) {
	message := &ProtobufMessage{
		Name: name,
	}

	err := VisitStructFields(structType, func(field *types.Var) error {
		fieldName := field.Name()
		if fieldName == "" {
			return fmt.Errorf("unexpected struct field with empty name")
		}

		fieldType, err := NewProtobufType(field.Type())
		if err != nil {
			return fmt.Errorf("failed to convert field %s of struct %s: %w", fieldName, name, err)
		}

		message.Fields = append(message.Fields, &ProtobufMessageField{
			Name: fieldName,
			Type: fieldType,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return message, nil
}

func convertSlice(slice *types.Slice) (*ProtobufArray, error) {
	element, err := NewProtobufType(slice.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to convert array element: %w", err)
	}

	return &ProtobufArray{
		Element: element,
	}, nil
}

func convertMap(mapType *types.Map) (*ProtobufMap, error) {
	keyType, err := NewProtobufType(mapType.Key())
	if err != nil {
		return nil, fmt.Errorf("failed to convert map key type: %w", err)
	}

	valueType, err := NewProtobufType(mapType.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to convert map value type: %w", err)
	}

	return &ProtobufMap{
		KeyType:   keyType,
		ValueType: valueType,
	}, nil
}
