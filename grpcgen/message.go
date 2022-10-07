package grpcgen

import (
	"fmt"
	"go/types"
)

type Message struct {
	Name   string
	Fields map[string]MessageType
}

func NewMessage(name string) *Message {
	return &Message{
		Name:   name,
		Fields: make(map[string]MessageType),
	}
}

type MessageType interface {
	fmt.Stringer

	Base() MessageType
	Package() *types.Package
}

type MessageTypeConverter struct {
	existingMessageTypes map[types.Type]MessageType
}

func NewMessageTypeConverter() *MessageTypeConverter {
	return &MessageTypeConverter{
		existingMessageTypes: make(map[types.Type]MessageType),
	}
}

func (converter *MessageTypeConverter) Convert(typ types.Type) MessageType {
	if existing, ok := converter.existingMessageTypes[typ]; ok {
		return existing
	}

	var result MessageType

	switch typ.(type) {
	case *types.Basic:
		result = NewBasicMessageType(typ.(*types.Basic))
	case *types.Named:
		named := typ.(*types.Named)
		underlying := named.Underlying()

		switch underlying.(type) {
		case *types.Struct:
			// TODO: structs to messages
			result = NewStructMessageType(named, underlying.(*types.Struct))
		case *types.Interface:
			// TODO: interface implementations
			result = NewInterfaceMessageType(named, underlying.(*types.Interface))
		case *types.Basic:
			// TODO: detect enums
			result = NewNamedMessageType(typ.(*types.Named))
		default:
			panic("unreachable")
		}

	case *types.Interface:
		interfaceType := typ.(*types.Interface)
		if interfaceType.Empty() {
			result = NewAnyMessageType()

			break
		}

		panic("unreachable")
	case *types.Pointer:
		result = NewPointerMessageType(converter, typ.(*types.Pointer))
	case *types.Slice:
		result = NewSliceMessageType(converter, typ.(*types.Slice))
	default:
		panic(fmt.Errorf("unsupported type: %T", typ))
	}

	converter.existingMessageTypes[typ] = result

	return result
}

type BasicMessageType struct {
	typ *types.Basic
}

func NewBasicMessageType(typ *types.Basic) *BasicMessageType {
	return &BasicMessageType{
		typ: typ,
	}
}

func (basic *BasicMessageType) String() string {
	return basic.typ.Name()
}

func (basic *BasicMessageType) Base() MessageType {
	return basic
}

func (basic *BasicMessageType) Package() *types.Package {
	return nil
}

type NamedMessageType struct {
	typ *types.Named
}

func NewNamedMessageType(typ *types.Named) *NamedMessageType {
	return &NamedMessageType{
		typ: typ,
	}
}

func (named *NamedMessageType) String() string {
	name := named.typ.String()

	pkg := named.Package()
	if pkg != nil {
		name = name[len(pkg.Name())+1:]
	}

	return name
}

func (named *NamedMessageType) Base() MessageType {
	return named
}

func (named *NamedMessageType) Package() *types.Package {
	return named.typ.Obj().Pkg()
}

type StructMessageType struct {
	named *types.Named
	typ   *types.Struct
}

func NewStructMessageType(named *types.Named, typ *types.Struct) *StructMessageType {
	return &StructMessageType{
		named: named,
		typ:   typ,
	}
}

func (structType *StructMessageType) Base() MessageType {
	return structType
}

func (structType *StructMessageType) Package() *types.Package {
	return structType.named.Obj().Pkg()
}

func (structType *StructMessageType) String() string {
	return structType.named.Obj().Name()
}

type InterfaceMessageType struct {
	named *types.Named
	typ   *types.Interface
}

func NewInterfaceMessageType(named *types.Named, typ *types.Interface) *InterfaceMessageType {
	fmt.Println("interface", typ.String())

	return &InterfaceMessageType{
		named: named,
		typ:   typ,
	}
}

func (interfaceType *InterfaceMessageType) Base() MessageType {
	return interfaceType
}

func (interfaceType *InterfaceMessageType) Package() *types.Package {
	fmt.Println(interfaceType.typ)

	return nil
}

func (interfaceType *InterfaceMessageType) String() string {
	return interfaceType.typ.String()
}

type AnyMessageType struct{}

func NewAnyMessageType() *AnyMessageType {
	return &AnyMessageType{}
}

func (any *AnyMessageType) Base() MessageType {
	return any
}

func (any *AnyMessageType) Package() *types.Package {
	return nil
}

func (any *AnyMessageType) String() string {
	return "google.protobuf.Any"
}

type PointerMessageType struct {
	typ  *types.Pointer
	base MessageType
}

func NewPointerMessageType(converter *MessageTypeConverter, typ *types.Pointer) *PointerMessageType {
	base := converter.Convert(typ.Elem())

	return &PointerMessageType{
		typ:  typ,
		base: base,
	}
}

func (pointer *PointerMessageType) Base() MessageType {
	return pointer.base
}

func (pointer *PointerMessageType) Package() *types.Package {
	return pointer.Package()
}

func (pointer *PointerMessageType) String() string {
	return pointer.base.String()
}

type SliceMessageType struct {
	typ  *types.Slice
	base MessageType
}

func NewSliceMessageType(converter *MessageTypeConverter, typ *types.Slice) *SliceMessageType {
	base := converter.Convert(typ.Elem())

	return &SliceMessageType{
		typ:  typ,
		base: base,
	}
}

func (pointer *SliceMessageType) Base() MessageType {
	return pointer.base
}

func (pointer *SliceMessageType) Package() *types.Package {
	return pointer.Package()
}

func (pointer *SliceMessageType) String() string {
	return pointer.base.String()
}
