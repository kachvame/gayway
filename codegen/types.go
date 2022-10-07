package codegen

import (
	"go/types"
)

func LookupType[T types.Type](typeScope *types.Scope, name string) (typeName *types.TypeName, result T) {
	typeObject := typeScope.Lookup(name)
	if typeObject == nil {
		return
	}

	var ok bool

	typeName, ok = typeObject.(*types.TypeName)
	if !ok {
		return
	}

	result = typeName.Type().(T)

	return
}
