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
	ndecls := 0
	var minPos, maxPos token.Pos
	i := 0
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
	}

	// Collect import specs from all package files.
	var imports []*ast.ImportSpec
	for _, f := range pkg.Syntax {
		imports = append(imports, f.Imports...)
	}

	// Collect declarations from all package files.
	var decls []ast.Decl
	if ndecls > 0 {
		decls = make([]ast.Decl, ndecls)
		i := 0 // current index
		n := 0 // number of filtered entries

		// Remove extra imports
		for decid, d := range decls {
			if dd, isGen := d.(*ast.GenDecl); isGen {
				if dd.Tok == token.IMPORT {
					decls[decid] = nil
					n++
				}
			}
		}

		// Add deduplicated imports to top of file
		specs := make([]ast.Spec, len(imports))
		for iid, imp := range imports {
			iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: imp.Path.Value}}
			specs[iid] = iSpec
		}

		decls = append([]ast.Decl{
			&ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: specs,
			},
		}, decls...)

		// Eliminate nil entries from the decls list if entries were
		// filtered. We do this using a 2nd pass in order to not disturb
		// the original declaration order in the source (otherwise, this
		// would also invalidate the monotonically increasing position
		// info within a single file).
		if n > 0 {
			i = 0
			for _, d := range decls {
				if d != nil {
					decls[i] = d
					i++
				}
			}
			decls = decls[0:i]
		}
	}

	f := &ast.File{
		Package:   0,
		Name:      ast.NewIdent(pkg.Name),
		Decls:     decls,
		FileStart: minPos,
		FileEnd:   maxPos,
		Imports:   imports,
	}

	return f
}
