package codegen

import "go/types"

func VisitExportedMethods(targetType *types.Named, visitor func(method *types.Func)) {
	for idx := 0; idx < targetType.NumMethods(); idx++ {
		method := targetType.Method(idx)
		if !method.Exported() {
			continue
		}

		visitor(method)
	}
}
