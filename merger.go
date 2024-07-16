package main

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

// MergePackageFiles creates a file AST by merging the ASTs of the
// files belonging to a package. The mode flags control merging behavior.
func MergePackageFiles(pkg *packages.Package) *ast.File {
	// Count the number of package docs, comments and declarations across
	// all package files. Also, compute sorted list of filenames, so that
	// subsequent iterations can always iterate in the same order.
	var minPos, maxPos token.Pos
	i := 0
	var pos token.Pos
	var ndecls = 0
	for _, f := range pkg.Syntax {
		pkg.Name = f.Name.Name
		i++
		ndecls += len(f.Decls)
		if i == 0 || f.FileStart < minPos {
			minPos = f.FileStart
		}
		if i == 0 || f.FileEnd > maxPos {
			maxPos = f.FileEnd
		}
		if f.Package > pos {
			// Keep the maximum package clause position as
			// position for the package clause of the merged
			// files.
			pos = f.Package
		}
	}

	// Collect import specs from all package files.
	var imports = []*ast.ImportSpec{}
	for _, f := range pkg.Syntax {
		for _, imp := range f.Imports {
			notDupe := true
			for _, imp2 := range imports {
				if imp.Path.Value == imp2.Path.Value {
					notDupe = false
				}
			}
			if notDupe {
				imports = append(imports, imp)
			}
		}
	}

	// Collect declarations from all package files.
	var decls = make([]ast.Decl, 0, ndecls)
	i = 0 // current index

	// Add declarations, excluding imports
	for _, f := range pkg.Syntax {
		for _, d := range f.Decls {
			if dd, isGen := d.(*ast.GenDecl); isGen {
				if dd.Tok == token.IMPORT || dd.Tok == token.COMMENT {
					continue
				}
			}
			decls = append(decls, d)
		}
	}

	// Add deduped imports to top of file
	if len(imports) > 0 {
		specs := make([]ast.Spec, len(imports))
		for iid, imp := range imports {
			specs[iid] = imp
		}
		decls = append([]ast.Decl{
			&ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: specs,
			},
		}, decls...)
	}

	f := &ast.File{
		Package:   pos,
		Name:      ast.NewIdent(pkg.Name),
		Decls:     decls,
		FileStart: minPos,
		FileEnd:   maxPos,
		Imports:   imports,
	}

	return f
}
