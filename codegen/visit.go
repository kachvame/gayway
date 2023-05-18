package codegen

import (
	"fmt"
	"go/types"
)

func VisitExportedMethods(targetType *types.Named, visitor func(method *types.Func) error) error {
	for idx := 0; idx < targetType.NumMethods(); idx++ {
		method := targetType.Method(idx)
		if !method.Exported() {
			continue
		}

		if err := visitor(method); err != nil {
			return fmt.Errorf("encountered error while visiting method %s: %v", method.Name(), err)
		}
	}

	return nil
}

func VisitTupleElements(tuple *types.Tuple, visitor func(idx int, variable *types.Var) error) error {
	for idx := 0; idx < tuple.Len(); idx++ {
		element := tuple.At(idx)

		if err := visitor(idx, element); err != nil {
			return fmt.Errorf("encountered error while visiting tuple element %d: %v", idx, err)
		}
	}

	return nil
}

func VisitStructFields(structType *types.Struct, visitor func(variable *types.Var) error) error {
	for idx := 0; idx < structType.NumFields(); idx++ {
		field := structType.Field(idx)
		if !field.Exported() {
			continue
		}

		if err := visitor(field); err != nil {
			return fmt.Errorf("encountered error while visiting struct field %d (%s): %v", idx, field.Name(), err)
		}
	}

	return nil
}
