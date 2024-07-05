package main

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/exp/maps"
)

var exclude = map[string]struct{}{"main": {}, "init": {}, "_": {}}

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

func findInterfaces(input string, typesPkgs []*types.Package, fset *token.FileSet, info *types.Info) map[*types.Interface]struct{} {
	imports := make(map[*types.Package]struct{})
	for _, pkg := range typesPkgs {
		// Get all interfaces so we don't mess up a type that should satisfy one of these interfaces
		getAllImports(pkg, imports)
	}
	return getAllPublicInterfaces(maps.Keys(imports))
}

func funcImplementsInterface(f *types.Func, interfaces map[*types.Interface]struct{}) bool {
	signature, ok := f.Type().(*types.Signature)
	if !ok {
		return false
	}
	if signature.Recv() == nil {
		return false
	}

	recvType := signature.Recv().Type()
	for i := range interfaces {
		if types.Implements(recvType, i) {
			return true
		}
	}
	return false
}

func createObfuscatedNames(info *types.Info, interfaces map[*types.Interface]struct{}) map[string]string {
	newNames := make(map[string]string)
	// Create obfuscated names for all identifiers in file
	// TODO: Check type switch statement
	for ident, obj := range info.Defs {
		if _, ok := exclude[ident.Name]; ok {
			continue
		}
		if f, ok := obj.(*types.Func); ok && funcImplementsInterface(f, interfaces) {
			continue
		}
		if obj == nil || obj.Pkg() == nil {
			continue
		}
		newNames[fullName(obj)] = nextName(ident.Name, true)
	}
	return newNames
}

func fullName(obj types.Object) string {
	if !obj.Exported() {
		return obj.Id()
	} else {
		return obj.Pkg().Name() + "." + obj.Id()
	}
}

func obfuscateAST(astPkgs []*ast.Package, info *types.Info, obfuscatedNames map[string]string) {
	for ident, obj := range info.Defs {
		if obj == nil || obj.Pkg() == nil {
			continue
		}
		if _, ok := obj.(*types.PkgName); ok {
			continue
		}
		if newName, ok := obfuscatedNames[fullName(obj)]; ok {
			ident.Name = newName
		}
	}
	for ident, obj := range info.Uses {
		if obj == nil || obj.Pkg() == nil {
			continue
		}
		if _, ok := obj.(*types.PkgName); ok {
			continue
		}
		if newName, ok := obfuscatedNames[fullName(obj)]; ok {
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

func obfuscate(input string, output string, fset *token.FileSet, astPkgs []*ast.Package, typesPkgs []*types.Package, info *types.Info) {
	interfaces := findInterfaces(input, typesPkgs, fset, info)
	obfuscatedNames := createObfuscatedNames(info, interfaces)
	obfuscateAST(astPkgs, info, obfuscatedNames)
	writeAST(input, output, astPkgs, fset)
}

func nextName(currentName string, isExported bool) string {
	return currentName + string(rune(rand.Intn(int('z'-'a'))+'a'))
}
