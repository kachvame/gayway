package codegen

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"unicode"
)

type StructDeclaration struct {
	Decl    *ast.GenDecl
	Spec    *ast.TypeSpec
	Struct  *ast.StructType
	Methods []*MethodDeclaration
}

type MethodDeclaration struct {
	Name string
	Decl *ast.FuncDecl
	Args []*Argument
}

type Argument struct {
	Name string
	Type *TypeExpression
}

type TypeExpression struct {
	Identifier *ast.Ident
	IsArray    bool
}

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

func FindStructs(pkg *packages.Package) StructDeclarations {
	// TODO: add debug logs
	res := make(StructDeclarations)

	// Find all structs
	for _, syn := range pkg.Syntax {
		for _, dec := range syn.Decls {
			if gen, ok := dec.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
				for _, spec := range gen.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := ts.Type.(*ast.StructType); ok {
							res[ts.Name.String()] = &StructDeclaration{
								Decl:   gen,
								Spec:   ts,
								Struct: structType,
							}
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

				args := make([]*Argument, 0, funcDecl.Type.Params.NumFields())
				for _, param := range funcDecl.Type.Params.List {
					paramType := unwrapExpr(param.Type)

					args = append(args, &Argument{
						Name: param.Names[0].Name,
						Type: paramType,
					})
				}

				res[structName].Methods = append(res[structName].Methods, &MethodDeclaration{
					Name: funcName,
					Decl: funcDecl,
					Args: args,
				})
			}
		}
	}

	return res
}
