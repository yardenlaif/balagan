package main

import (
	"go/ast"
	"go/printer"
	"os"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// type fake struct{}

// func (f *fake) AFact() {}

// var SymbolWriter = analysis.Analyzer{
// 	Name:      "SymbolWriter",
// 	Doc:       "Obfuscate all symbols",
// 	FactTypes: []analysis.Fact{analysis.Fact(&fake{})},
// 	Run: func(p *analysis.Pass) (interface{}, error) {
// 		for _, file := range p.Files {
// 			var root ast.Node
// 			ast.Inspect(file, func(n ast.Node) bool {
// 				root = n
// 				return false
// 			})

// 			astutil.Apply(root, func(c *astutil.Cursor) bool {
// 				// if o, ok := c.Node().(*ast.Ident); ok {
// 				// updateSymbol(pass.TypesInfo.ObjectOf(o), o)
// 				// }
// 				// if comment, ok := c.Node().(*ast.Comment); ok {
// 				// 	comment.Text = "//"
// 				// 	return false
// 				// }
// 				// if doc, ok := c.Node().(*ast.

// 				return true
// 			}, nil)

// 			// outFile, err := os.Create(p.Fset.File(file.Pos()).Name())
// 			// if err != nil {
// 			// 	panic("err")
// 			// }
// 			// printer.Fprint(outFile, p.Fset, file)
// 			// outFile.Close()
// 			if strings.Contains(p.Fset.File(file.Pos()).Name(), "balagan") {
// 				printer.Fprint(os.Stdout, p.Fset, file)
// 			}
// 		}
// 		return nil, nil
// 	},
// 	Requires: []*analysis.Analyzer{&SymbolObfuscator},
// }

func inPlaceRename(pkgs []*packages.Package) (interface{}, error) {
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			var root ast.Node
			ast.Inspect(file, func(n ast.Node) bool {
				root = n
				return false
			})

			astutil.Apply(root, func(c *astutil.Cursor) bool {
				// if o, ok := c.Node().(*ast.Ident); ok {
				// updateSymbol(pass.TypesInfo.ObjectOf(o), o)
				// }
				// if comment, ok := c.Node().(*ast.Comment); ok {
				// 	comment.Text = "//"
				// 	return false
				// }
				// if doc, ok := c.Node().(*ast.

				return true
			}, nil)

			filename := pkg.Fset.File(file.Pos()).Name()
			outFile, err := os.Create(filename)
			if err != nil {
				panic("err")
			}
			printer.Fprint(outFile, pkg.Fset, file)
			outFile.Close()
			// printer.Fprint(os.Stdout, pkg.Fset, file)
		}
	}
	return nil, nil
}
