package codegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
	"unicode"
)

type StructDeclaration struct {
	Decl    *ast.GenDecl
	Spec    *ast.TypeSpec
	Struct  *ast.StructType
	Fields  []*Value
	Methods []*MethodDeclaration
}

type MethodDeclaration struct {
	Name    string
	Decl    *ast.FuncDecl
	Args    []*Value
	Results []*Value
}

type Value struct {
	Name string
	Type *TypeExpression
}

type TypeExpression struct {
	Identifier *ast.Ident
	IsArray    bool
}

type Enums map[string][]*types.Const

type StructDeclarations map[string]*StructDeclaration

func isExported(name string) bool {
	return unicode.IsUpper([]rune(name)[0])
}

func unwrapExpr(expr ast.Expr) *TypeExpression {
	var unwrapExprHelper func(expr ast.Expr) (*ast.Ident, bool)
	unwrapExprHelper = func(expr ast.Expr) (*ast.Ident, bool) {
		switch expr.(type) {
		case *ast.Ident:
			return expr.(*ast.Ident), false
		case *ast.StarExpr:
			starExpr := expr.(*ast.StarExpr)

			return unwrapExprHelper(starExpr.X)
		case *ast.ArrayType:
			arrayType := expr.(*ast.ArrayType)
			elementType, _ := unwrapExprHelper(arrayType.Elt)

			return elementType, true
		}

		return nil, false
	}

	identifier, isArray := unwrapExprHelper(expr)

	return &TypeExpression{
		Identifier: identifier,
		IsArray:    isArray,
	}
}

func parseFields(fields *ast.FieldList, defaultName func(index int, typ *TypeExpression) string) []*Value {
	if fields.NumFields() <= 0 {
		return nil
	}

	values := make([]*Value, 0, fields.NumFields())
	for index, field := range fields.List {
		paramType := unwrapExpr(field.Type)

		var name string
		if len(field.Names) > 0 {
			name = field.Names[0].Name
		} else {
			name = defaultName(index, paramType)
		}

		values = append(values, &Value{
			Name: name,
			Type: paramType,
		})
	}

	return values
}

func FindEnums(pkg *packages.Package) Enums {
	typePrefix := fmt.Sprintf("%s.", pkg.ID)
	typesScope := pkg.Types.Scope()

	enumValues := make(Enums)
	for _, syn := range pkg.Syntax {
		for _, dec := range syn.Decls {
			if gen, ok := dec.(*ast.GenDecl); ok {
				for _, spec := range gen.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							object := typesScope.Lookup(name.Name)
							if constant, ok := object.(*types.Const); ok {
								typ := constant.Type().String()

								if !strings.HasPrefix(typ, typePrefix) {
									continue
								}

								typ = strings.TrimPrefix(typ, typePrefix)

								enumValues[typ] = append(enumValues[typ], constant)
							}
						}
					}
				}
			}
		}
	}

	return enumValues
}

func FindStructs(pkg *packages.Package, fieldlessStructs map[string]struct{}) StructDeclarations {
	// TODO: add debug logs
	res := make(StructDeclarations)

	structFieldDefaultName := func(_ int, typ *TypeExpression) string {
		return typ.Identifier.Name
	}

	valueDefaultName := func(index int, _ *TypeExpression) string {
		return fmt.Sprintf("Value%d", index)
	}

	// Find all structs
	for _, syn := range pkg.Syntax {
		for _, dec := range syn.Decls {
			if gen, ok := dec.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
				for _, spec := range gen.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := ts.Type.(*ast.StructType); ok {
							structName := ts.Name.String()
							if !isExported(structName) {
								continue
							}

							_, isFieldless := fieldlessStructs[structName]

							structDeclaration := &StructDeclaration{
								Decl:   gen,
								Spec:   ts,
								Struct: structType,
							}

							if !isFieldless {
								structDeclaration.Fields = parseFields(structType.Fields, structFieldDefaultName)
							}

							res[structName] = structDeclaration
						}
					}
				}
			}
		}
	}

	// Find all struct methods
	for _, syn := range pkg.Syntax {
		for _, dec := range syn.Decls {
			if funcDecl, ok := dec.(*ast.FuncDecl); ok {
				if funcDecl.Recv == nil {
					continue
				}

				funcName := funcDecl.Name.String()
				if !isExported(funcName) {
					continue
				}

				receiver := funcDecl.Recv.List[0]
				receiverExpr := receiver.Type

				identifier := unwrapExpr(receiverExpr)

				structName := identifier.Identifier.Name
				if _, ok := res[structName]; !ok {
					continue
				}

				res[structName].Methods = append(res[structName].Methods, &MethodDeclaration{
					Name:    funcName,
					Decl:    funcDecl,
					Args:    parseFields(funcDecl.Type.Params, valueDefaultName),
					Results: parseFields(funcDecl.Type.Results, valueDefaultName),
				})
			}
		}
	}

	return res
}
