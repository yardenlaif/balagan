package main

import (
	"go/types"

	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/packages"
)

var found = make(map[string]struct{})

func findInterfacesInPkg(pkg *packages.Package, interfaces map[*types.Interface]struct{}) {
	if _, ok := found[pkg.PkgPath]; ok {
		return
	}
	found[pkg.PkgPath] = struct{}{}
	for _, importPkg := range pkg.Imports {
		findInterfacesInPkg(importPkg, interfaces)
	}

	for _, typeAndValue := range pkg.TypesInfo.Types {
		t := typeAndValue.Type
		if named, ok := t.(*types.Named); ok {
			t = named.Underlying()
		}
		if n, ok := t.(*types.Interface); ok {
			interfaces[n] = struct{}{}
		}
	}
}

func findInterfaces(pkgs []*packages.Package) []*types.Interface {
	interfaces := make(map[*types.Interface]struct{})
	for _, pkg := range pkgs {
		findInterfacesInPkg(pkg, interfaces)
	}
	return maps.Keys(interfaces)
}

func findImplementationsInPkg(pkg *packages.Package, interfaces []*types.Interface, implementations map[*types.Func][]*types.Func) {
	if _, ok := found[pkg.PkgPath]; ok {
		return
	}
	found[pkg.PkgPath] = struct{}{}
	for _, importPkg := range pkg.Imports {
		findImplementationsInPkg(importPkg, interfaces, implementations)
	}

	for _, def := range pkg.TypesInfo.Defs {
		if fun, ok := def.(*types.Func); ok {
			recv := fun.Type().(*types.Signature).Recv()
			if recv == nil {
				continue
			}

			for _, i := range interfaces {
				if !types.Implements(recv.Type(), i) {
					continue
				}

				obj, _, _ := types.LookupFieldOrMethod(i, false, recv.Pkg(), fun.Name())
				if method, ok := obj.(*types.Func); ok {
					if method == fun {
						continue
					}
					implementations[method] = append(implementations[method], fun)
				}
			}
		}
	}
}

func findImplementations(pkgs []*packages.Package) map[*types.Func][]*types.Func {
	implementations := make(map[*types.Func][]*types.Func)
	interfaces := findInterfaces(pkgs)
	found = make(map[string]struct{})
	for _, pkg := range pkgs {
		findImplementationsInPkg(pkg, interfaces, implementations)
	}
	return implementations
}
