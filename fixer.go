package main

import (
	"fmt"
	"go/ast"
	"go/printer"
	"io"
	"os"
	"path"
	"strings"

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

func inPlaceRename(pkgs []*packages.Package, oldPath string, newPath string) (interface{}, error) {
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
			filename = strings.Replace(filename, oldPath, newPath, 1)
			dirname := path.Dir(filename)
			err := os.MkdirAll(dirname, 0777)
			if err != nil {
				panic(err)
			}
			outFile, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			printer.Fprint(outFile, pkg.Fset, file)
			outFile.Close()
			// printer.Fprint(os.Stdout, pkg.Fset, file)
		}
		for _, filename := range pkg.OtherFiles {
			fmt.Printf("Copying: %s\n", filename)
			origFile, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			defer origFile.Close()
			filename = strings.Replace(filename, oldPath, newPath, 1)
			dirname := path.Dir(filename)
			err = os.MkdirAll(dirname, 0777)
			if err != nil {
				panic(err)
			}
			newFile, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			defer newFile.Close()
			io.Copy(newFile, origFile)
		}
	}
	return nil, nil
}
