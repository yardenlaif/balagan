package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"os"
	"path"

	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

func findGoModFile(dir string) string {
	goMod := path.Join(dir, "go.mod")
	_, err := os.Open(goMod)
	if os.IsNotExist(err) {
		parent, _ := path.Split(dir)
		if parent == "" {
			panic("No go mod file found")
		}
		return findGoModFile(parent)
	}
	return goMod
}

func getModulePath(dir string) string {
	goModFile := findGoModFile(dir)
	log.Println("goModFile: ", goModFile)
	goMod, err := os.ReadFile(goModFile)
	if err != nil {
		log.Fatalln(err)
	}
	m, err := modfile.Parse(goModFile, goMod, nil)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("mod: %v\n", m.Module.Mod.Path)
	return m.Module.Mod.Path
}

var filled = make(map[string]struct{})

func fillInfo(pkg *packages.Package, info *types.Info) {
	if _, ok := filled[pkg.Name]; ok {
		return
	}
	filled[pkg.Name] = struct{}{}
	log.Printf("filling: %v", pkg.Name)
	for _, importPkg := range pkg.Imports {
		fillInfo(importPkg, info)
	}
	maps.Copy(info.Defs, pkg.TypesInfo.Defs)
	maps.Copy(info.Uses, pkg.TypesInfo.Uses)
	for k, v := range pkg.TypesInfo.Types {
		info.Types[k] = v
	}
	for k, v := range pkg.TypesInfo.Scopes {
		info.Scopes[k] = v
	}
}

func checkDir(path string) ([]*packages.Package, *types.Info) {
	info := &types.Info{
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Scopes: make(map[ast.Node]*types.Scope),
	}
	var allPkgs []*packages.Package
	pkgConf := packages.Config{Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedDeps | packages.NeedImports | packages.NeedTypesInfo | packages.NeedName | packages.NeedModule}
	pkgConf.Dir = path
	// TODO: Document adding ... to obfuscate recursively
	pkgs, err := packages.Load(&pkgConf, path+"/...")
	if err != nil {
		log.Fatal(err)
	}

	allPkgs = append(allPkgs, pkgs...)
	for _, pkg := range pkgs {
		fillInfo(pkg, info)
	}
	return allPkgs, info
}

func main() {
	// fset := token.NewFileSet()

	// // import "go/types" and "go/importer"
	// conf := types.Config{Importer: importer.ForCompiler(fset, "gc", nil)}

	// // types.TypeOf() requires all three maps are populated

	pkgs, info := checkDir(os.Args[1])
	implementations := findImplementations(pkgs)
	noRename := make(map[types.Object]struct{})
	for f, impls := range implementations {
		for _, impl := range impls {
			noRename[impl] = struct{}{}
		}
		noRename[f] = struct{}{}
	}
	symbols := make(map[interface{}]string)
	obfuscated := make(map[string]struct{})
	for _, pkg := range pkgs {
		for _, importPkg := range pkg.Imports {
			if importPkg.Module == pkg.Module {
				if _, ok := obfuscated[importPkg.PkgPath]; !ok {
					p := NewPackageObfuscator(importPkg, info, symbols, noRename)
					p.Obfuscate()
					obfuscated[importPkg.PkgPath] = struct{}{}
				}
			}
		}
	}
	for _, pkg := range pkgs {
		if _, ok := obfuscated[pkg.PkgPath]; !ok {
			p := NewPackageObfuscator(pkg, info, symbols, noRename)
			p.Obfuscate()
			obfuscated[pkg.PkgPath] = struct{}{}
		}
	}
	inPlaceRename(pkgs)
	// spew.Dump(findImplementations(info))
	// spew.Printf("interfaces: %v\n", findInterfaces(info))
	// spew.Printf("implementations: %v\n", findImplementations(info))
	// renameSymbols(info, pkgs)
	// inPlaceRename(fset, pkgs)
	// singlechecker.Main(&SymbolWriter)
	// pass := analysis.Pass{TypesInfo: info, Files: []*ast.File{f}, Pkg: pkg, Fset: fset}

	// anchorRun(&pass)
	// implementationFinderRun(&pass)
	// run(&pass)
	// inPlaceRename(&pass)

	// singlechecker.Main(InPlaceRenamer)
}
