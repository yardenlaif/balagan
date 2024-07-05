package main

import (
	"go/ast"
	"go/types"
	"log"
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
		log.Printf("Not obfuscating: %v\n", i)
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

	if assign, ok := i.Obj.Decl.(*ast.AssignStmt); ok {
		for _, lhs := range assign.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok {
				ident.Name = newName
			}
		}
	}
}

func (p *PackageObfuscator) shouldCreateObfuscation(i *ast.Ident) bool {
	o := p.typesInfo.ObjectOf(i)
	return o != nil && o.Pkg() != nil && o.Pkg().Path() == p.packagePath
}

func (p *PackageObfuscator) obfuscationKey(i *ast.Ident) (interface{}, bool) {
	if i.Obj != nil && i.Obj.Decl != nil {
		if assign, ok := i.Obj.Decl.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 {
			if _, ok := assign.Rhs[0].(*ast.TypeAssertExpr); ok {
				return i.Obj.Decl, true
			}
		}
	}
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
