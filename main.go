package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"log"
	"path"
	"path/filepath"

	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func getModulePath(dir string) string {
	// goMod, err := os.ReadFile(path + "/go.mod")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// m, err := modfile.Parse(path+"/go.mod", goMod, nil)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Printf("mod: %v\n", m.Module.Mod.Path)
	// return m.Module.Mod.Path
	return path.Join("github.com/Rookout/GoSDK", dir)
}

func checkDir(conf *types.Config, info *types.Info, path string, fset *token.FileSet) map[string]*ast.Package {
	allPkgs := make(map[string]*ast.Package)
	filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		if info == nil || info.Type() != fs.ModeDir {
			return nil
		}

		pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		for _, pkg := range pkgs {
			allPkgs[path] = pkg
		}

		return nil
	})
	for path, pkg := range allPkgs {
		fmt.Printf("path: %v\n", path)
		_, err := conf.Check(path, fset, maps.Values(pkg.Files), info)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return allPkgs
}

func main() {
	// fset := token.NewFileSet()

	// // import "go/types" and "go/importer"
	// conf := types.Config{Importer: importer.Default()}
	// conf.Error = func(err error) { fmt.Printf("err: %v\n", err) }

	// // types.TypeOf() requires all three maps are populated
	// info := &types.Info{
	// 	Defs:   make(map[*ast.Ident]types.Object),
	// 	Uses:   make(map[*ast.Ident]types.Object),
	// 	Types:  make(map[ast.Expr]types.TypeAndValue),
	// 	Scopes: make(map[ast.Node]*types.Scope),
	// }

	// pkgs := checkDir(&conf, info, os.Args[1], fset)
	// fmt.Printf("interfaces: %v\n", findInterfaces(info))
	// fmt.Printf("implementations: %v\n", findImplementations(info))
	// renameSymbols(info, pkgs)
	// inPlaceRename(fset, pkgs)
	singlechecker.Main(&SymbolWriter)
	// pass := analysis.Pass{TypesInfo: info, Files: []*ast.File{f}, Pkg: pkg, Fset: fset}

	// anchorRun(&pass)
	// implementationFinderRun(&pass)
	// run(&pass)
	// inPlaceRename(&pass)

	// singlechecker.Main(InPlaceRenamer)
}
