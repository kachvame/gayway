package codegen

import (
	"fmt"
	"go.uber.org/multierr"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
)

const (
	discordgoPackageName = "github.com/bwmarrin/discordgo"
	discordgoSession     = "Session"
)

func (generator *Generator) loadDiscordgoPackage() (*packages.Package, error) {
	packageConfig := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}

	discordgoPackages, err := packages.Load(packageConfig, discordgoPackageName)
	if err != nil {
		return nil, fmt.Errorf("failed to find discordgo package: %w", err)
	}

	packages.Visit(discordgoPackages, nil, func(pkg *packages.Package) {
		for _, packageErr := range pkg.Errors {
			multierr.AppendInto(&err, packageErr)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("encountered errors while loading discordgo package: %w", err)
	}

	discordgoPackage := discordgoPackages[0]

	return discordgoPackage, nil
}

func (generator *Generator) findSessionStruct(discordgoPackage *packages.Package) (*types.Named, error) {
	typeScope := discordgoPackage.Types.Scope()
	typeObject := typeScope.Lookup(discordgoSession)
	if typeObject == nil {
		return nil, fmt.Errorf("failed to find discordgo Session struct type")
	}

	typeName, ok := typeObject.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("failed to cast Session type object to type name")
	}

	sessionType, ok := typeName.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("failed to cast Session type to named type")
	}

	return sessionType, nil
}
