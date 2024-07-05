package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type PackageObfuscator struct {
	symbols             map[interface{}]string
	currentName         string
	currentExportedName string
	packagePath         string
	noRename            map[types.Object]struct{}
	typesInfo           *types.Info
	files               []*ast.File
}

func NewPackageObfuscator(pkg *packages.Package, info *types.Info, symbols map[interface{}]string, noRename map[types.Object]struct{}) PackageObfuscator {
	return PackageObfuscator{
		currentName:         "a",
		currentExportedName: "A",
		packagePath:         pkg.PkgPath,
		noRename:            noRename,
		symbols:             symbols,
		typesInfo:           info,
		files:               pkg.Syntax,
	}
}

func (p *PackageObfuscator) Obfuscate() {
	for _, file := range p.files {
		var root ast.Node
		ast.Inspect(file, func(n ast.Node) bool {
			root = n
			return false
		})

		astutil.Apply(root, func(c *astutil.Cursor) bool {
			node := c.Node()
			switch t := node.(type) {
			case *ast.Ident:
				p.obfuscateSymbol(t)
			case *ast.TypeSwitchStmt:
				assign := t.Assign.(*ast.AssignStmt)
				fmt.Printf("assign: %v\n", assign)
				astutil.Apply(node, func(c *astutil.Cursor) bool {
					switch c.Node().(type) {
					case *ast.TypeAssertExpr:
						fmt.Printf("parent: %v %T\n", c.Parent(), c.Parent())
					}
					return true
				}, nil)
				return false
			}
			return true
		}, nil)
	}
}

func (p *PackageObfuscator) obfuscateSymbol(i *ast.Ident) {
	key, ok := p.obfuscationKey(i)
	if !ok {
		return
	}
	if key == nil {
		return
	}
	if newName, ok := p.symbols[key]; ok {
		i.Name = newName
		return
	}
	if !p.shouldCreateObfuscation(i) {
		return
	}
	name := i.Name
	var newName string
	if unicode.IsUpper(rune(name[0])) {
		newName = p.currentExportedName
		p.currentExportedName = nextName(p.currentExportedName, true)
	} else {
		newName = p.currentName
		p.currentName = nextName(p.currentName, false)
	}

	p.symbols[key] = newName
	i.Name = newName
}

func isMainFunc(i *ast.Ident, o types.Object) bool {
	if _, ok := o.Type().(*types.Signature); ok {
		return i.Name == "main"
	}
	return false
}

func (p *PackageObfuscator) shouldCreateObfuscation(i *ast.Ident) bool {
	o := p.typesInfo.ObjectOf(i)
	if isMainFunc(i, o) {
		return false
	}
	return o != nil && o.Pkg() != nil && o.Pkg().Path() == p.packagePath
}

func (p *PackageObfuscator) obfuscationKey(i *ast.Ident) (interface{}, bool) {
	o := p.typesInfo.ObjectOf(i)
	if o == nil {
		return nil, false
	}
	if basic, ok := o.Type().(*types.Basic); ok {
		if basic.Kind() == types.Invalid {
			return nil, false
		}
	}
	if _, ok := p.noRename[o]; ok {
		return nil, false
	}
	// Don't rename unnamed objects
	if o.Name() == "" || o.Name() == "_" {
		return nil, false
	}
	return o, true
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
