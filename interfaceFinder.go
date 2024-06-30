package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/exp/maps"
)

func getPublicInterfaces(pkg *types.Package) map[*types.Interface]struct{} {
	var interfaces map[*types.Interface]struct{}
	for _, name := range pkg.Scope().Names() {
		typeName, ok := pkg.Scope().Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		t := typeName.Type()
		if _, ok := t.(*types.Named); ok {
			t = t.Underlying()
		}
		if i, ok := t.(*types.Interface); ok {
			interfaces[i] = struct{}{}
			log.Printf("Found interface: %+v\n", i)
		}
	}
	return interfaces
}

func findInterfaces(path string) map[*types.Interface]struct{} {
	interfaces := make(map[*types.Interface]struct{})
	var fset token.FileSet
	conf := &types.Config{Importer: importer.Default()}
	info := types.Info{
		Scopes: make(map[ast.Node]*types.Scope),
	}

	astPkgs, _ := parser.ParseDir(&fset, path, nil, 0)
	for _, astPkg := range astPkgs {
		typesPkg, _ := conf.Check(path, &fset, maps.Values(astPkg.Files), &info)
		maps.Copy(interfaces, getPublicInterfaces(typesPkg))
		for _, imp := range typesPkg.Imports() {
			maps.Copy(interfaces, getPublicInterfaces(imp))
		}
	}
	return interfaces
}
