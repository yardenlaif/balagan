package main

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

type PackageObfuscator struct {
	symbols             map[types.Object]string
	currentName         string
	currentExportedName string
	packagePath         string
	noRename            map[types.Object]struct{}
}

type symbolMap struct {
	m map[types.Object]string
}

func (s *symbolMap) AFact() {}

var SymbolObfuscator = analysis.Analyzer{
	Name:      "SymbolObfuscator",
	Doc:       "Obfuscate all symbols",
	FactTypes: []analysis.Fact{&symbolMap{}},
	Run: func(p *analysis.Pass) (interface{}, error) {
		implementations := p.ResultOf[&InterfaceImplementationFinder].(*implementationMap)
		symbols := make(map[types.Object]string)
		renamePkgSymbols(symbols, p.TypesInfo, p.Pkg.Path(), p.Files, *implementations)
		for _, i := range p.Pkg.Imports() {
			var fact symbolMap
			p.ImportPackageFact(i, &fact)
			for k, v := range fact.m {
				symbols[k] = v
			}
		}

		p.ExportPackageFact(&symbolMap{m: symbols})
		return nil, nil
	},
	Requires: []*analysis.Analyzer{&InterfaceImplementationFinder},
}

func renamePkgSymbols(symbols map[types.Object]string, info *types.Info, pkgName string, files []*ast.File, implementations map[*types.Func][]*types.Func) {
	p := PackageObfuscator{symbols: symbols, currentName: "a", currentExportedName: "A", packagePath: pkgName, noRename: make(map[types.Object]struct{})}
	for f, impls := range implementations {
		for _, impl := range impls {
			p.noRename[impl] = struct{}{}
		}
		p.noRename[f] = struct{}{}
	}
	for _, file := range files {
		var root ast.Node
		ast.Inspect(file, func(n ast.Node) bool {
			root = n
			return false
		})

		astutil.Apply(root, func(c *astutil.Cursor) bool {
			node := c.Node()
			switch t := node.(type) {
			case *ast.Ident:
				p.renameSymbol(info.ObjectOf(t), t)
			}
			return true
		}, nil)
	}
}

func (p *PackageObfuscator) renameSymbol(o types.Object, i *ast.Ident) {
	if o == nil || o.Pkg() == nil {
		return
	}

	if _, ok := p.noRename[o]; ok {
		return
	}

	if newName, ok := p.symbols[o]; ok {
		i.Name = newName
		return
	}

	// Pkg is nil if o is part of the language (for example - the string type)
	if o.Pkg() == nil || o.Pkg().Path() != p.packagePath {
		return
	}

	// Don't rename unnamed objects
	if o.Name() == "" || o.Name() == "_" {
		return
	}

	if basic, ok := o.Type().(*types.Basic); ok {
		if basic.Kind() == types.Invalid {
			return
		}
	}

	name := o.Name()
	var newName string
	if unicode.IsUpper(rune(name[0])) {
		newName = p.currentExportedName
		p.currentExportedName = nextName(p.currentExportedName, true)
	} else {
		newName = p.currentName
		p.currentName = nextName(p.currentName, false)
	}

	p.symbols[o] = newName
	i.Name = newName
}

func nextName(currentName string, isExported bool) string {
	newName := ""
	for i, c := range currentName {
		if unicode.IsLetter(rune(c + 1)) {
			newName += string(c+1) + currentName[i+1:]
			return newName
		}
		newName += string(c)
	}

	newName = "a"
	if isExported {
		newName = "A"
	}
	newName += strings.Repeat("a", len(currentName))
	return newName
}
