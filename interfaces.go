package main

import (
	"go/types"

	"golang.org/x/exp/maps"
)

func getAllImports(pkg *types.Package, imports map[*types.Package]struct{}) {
	if _, ok := imports[pkg]; ok {
		return
	}
	imports[pkg] = struct{}{}

	for _, imp := range pkg.Imports() {
		getAllImports(imp, imports)
	}
}

func getPublicInterfaces(pkg *types.Package) map[*types.Interface]struct{} {
	interfaces := make(map[*types.Interface]struct{})
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
		}
	}
	return interfaces
}

func getAllPublicInterfaces(pkgs []*types.Package) map[*types.Interface]struct{} {
	interfaces := make(map[*types.Interface]struct{})
	for _, pkg := range pkgs {
		maps.Copy(interfaces, getPublicInterfaces(pkg))
	}
	return interfaces
}

func findInterfaces(typesPkgs []*types.Package) map[*types.Interface]struct{} {
	imports := make(map[*types.Package]struct{})
	for _, pkg := range typesPkgs {
		// Get all interfaces so we don't mess up a type that should satisfy one of these interfaces
		getAllImports(pkg, imports)
	}
	return getAllPublicInterfaces(maps.Keys(imports))
}
