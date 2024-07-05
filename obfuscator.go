package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/exp/maps"
)

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
			log.Printf("Found interface: %+v\n", i)
		}
	}
	return interfaces
}

func getPublicInterfacesRecursively(pkg *types.Package) map[*types.Interface]struct{} {
	interfaces := getPublicInterfaces(pkg)
	for _, imp := range pkg.Imports() {
		maps.Copy(interfaces, getPublicInterfacesRecursively(imp))
	}
	return interfaces
}

func findInterfaces(input string, astPkgs []*ast.Package, fset *token.FileSet, info *types.Info) map[*types.Interface]struct{} {
	interfaces := make(map[*types.Interface]struct{})
	conf := &types.Config{Importer: importer.Default()}
	for _, astPkg := range astPkgs {
		// Get all interfaces so we don't mess up a type that should satisfy one of these interfaces
		pkg, _ := conf.Check(input, fset, maps.Values(astPkg.Files), info)
		maps.Copy(interfaces, getPublicInterfacesRecursively(pkg))
	}
	return interfaces
}

func createObfuscatedNames(info *types.Info, interfaces map[*types.Interface]struct{}) map[types.Object]string {
	newNames := make(map[types.Object]string)
	// Create obfuscated names for all identifiers in file
	// TODO: Check type switch statement
	for ident, obj := range info.Defs {
		if ident.Name == "main" {
			continue
		}
		newNames[obj] = nextName(ident.Name, true)
	}
	return newNames
}

func obfuscateAST(astPkgs []*ast.Package, info *types.Info, obfuscatedNames map[types.Object]string) {
	for ident, obj := range info.Defs {
		if newName, ok := obfuscatedNames[obj]; ok {
			ident.Name = newName
		}
	}
	for ident, obj := range info.Uses {
		if newName, ok := obfuscatedNames[obj]; ok {
			ident.Name = newName
		}
	}
}

func writeAST(input string, output string, astPkgs []*ast.Package, fset *token.FileSet) {
	// TODO: Do something with errors returned from Abs
	// TODO: Make sure output is empty and writable before doing obfuscation
	input, _ = filepath.Abs(input)
	output, _ = filepath.Abs(output)
	fmt.Printf("input: %s, output: %s\n", input, output)
	for _, astPkg := range astPkgs {
		for _, file := range astPkg.Files {
			filename, _ := filepath.Abs(fset.File(file.Pos()).Name())
			filename = strings.Replace(filename, input, output, 1)
			dirname := path.Dir(filename)
			// TODO: Fix dir permissions
			err := os.MkdirAll(dirname, 0777)
			if err != nil {
				panic(err)
			}
			outFile, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			printer.Fprint(outFile, fset, file)
			outFile.Close()
		}
	}
}

func obfuscate(input string, fset *token.FileSet, astPkgs []*ast.Package) map[*types.Interface]struct{} {
	info := types.Info{
		Scopes: make(map[ast.Node]*types.Scope),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
	}

	interfaces := findInterfaces(input, astPkgs, fset, &info)
	obfuscatedNames := createObfuscatedNames(&info, interfaces)
	obfuscateAST(astPkgs, &info, obfuscatedNames)
	writeAST(input, "./output", astPkgs, fset)

	for _, astPkg := range astPkgs {
		for _, file := range astPkg.Files {
			printer.Fprint(os.Stdout, fset, file)
		}
	}
	return interfaces
}

func nextName(currentName string, isExported bool) string {
	return currentName + "1"
}
