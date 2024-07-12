package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

// Don't obfuscate these names as they hold semantic meaning
var exclude = map[string]struct{}{"main": {}, "init": {}, "_": {}}

type Obfuscator struct {
	interfaces      map[*types.Interface]struct{}
	info            *types.Info
	currentName     int
	obfuscatedNames map[string]string
	astPkgs         []*ast.Package
	fset            *token.FileSet
	sourcePath      string
	targetPath      string
}

func NewObfuscator(sourcePath string, targetPath string, ignorePaths []string) (*Obfuscator, error) {
	o := &Obfuscator{
		sourcePath:      sourcePath,
		targetPath:      targetPath,
		fset:            token.NewFileSet(),
		obfuscatedNames: make(map[string]string),
	}

	// This is necessary for the importer to work
	err := os.Chdir(sourcePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "Unable to change directory to source directory %s", sourcePath)
	}

	conf := &types.Config{Importer: importer.ForCompiler(o.fset, "source", nil), Error: func(err error) {}}
	o.info = &types.Info{
		Scopes: make(map[ast.Node]*types.Scope),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
	}
	astPkgs := make(map[*ast.Package]struct{})
	typesPkgs := make(map[*types.Package]struct{})

	err = filepath.WalkDir(sourcePath, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		for _, ignore := range ignorePaths {
			if strings.HasPrefix(path, ignore) {
				return nil
			}
		}

		return o.parseDir(path, e, astPkgs, typesPkgs, conf)
	})
	if err != nil {
		return nil, errors.WithMessagef(err, "Unable to recurse source directory %s", sourcePath)
	}

	o.interfaces = findInterfaces(maps.Keys(typesPkgs))
	o.astPkgs = maps.Keys(astPkgs)
	return o, nil
}

func (o *Obfuscator) parseDir(path string, e fs.DirEntry, astPkgs map[*ast.Package]struct{}, typesPkgs map[*types.Package]struct{}, conf *types.Config) error {
	// Only parse directories
	if e.Type() != fs.ModeDir {
		return nil
	}

	fmt.Printf("Parsing %-50s    ", path)
	start := time.Now()
	pkgs, err := parser.ParseDir(o.fset, path, nil, parser.ParseComments)
	if err != nil {
		errors.WithMessagef(err, "Unable to parse source directory %s to AST", path)
	}

	for _, astPkg := range pkgs {
		if _, ok := astPkgs[astPkg]; ok {
			continue
		}
		astPkgs[astPkg] = struct{}{}

		// Ignore this error, as it may be caused when files that would usually not be compiled in (because of build tags) are included in the type check. Also, it is the user's responsibility to check their code.
		pkg, _ := conf.Check(path, o.fset, maps.Values(astPkg.Files), o.info)
		typesPkgs[pkg] = struct{}{}
	}
	fmt.Printf("%s\n", time.Since(start))
	return nil
}

func (o *Obfuscator) Obfuscate() error {
	fmt.Println("Obfuscating...")
	o.createObfuscatedNames()
	o.obfuscateAST()
	removeComments(o.astPkgs)
	return o.writeAST()
}

func (o *Obfuscator) funcImplementsInterface(f *types.Func) bool {
	signature, ok := f.Type().(*types.Signature)
	if !ok {
		return false
	}
	if signature.Recv() == nil {
		return false
	}

	recvType := signature.Recv().Type()
	for i := range o.interfaces {
		if types.Implements(recvType, i) {
			return true
		}
	}
	return false
}

func (o *Obfuscator) createObfuscatedNames() {
	for ident, obj := range o.info.Defs {
		if _, ok := exclude[ident.Name]; ok {
			continue
		}
		// If this function implements a public interface, don't change it's name so we don't break the interface implementation
		if f, ok := obj.(*types.Func); ok && o.funcImplementsInterface(f) {
			continue
		}
		// Don't obfuscate Universe objects
		if obj == nil || obj.Pkg() == nil {
			continue
		}
		// This can happen when we have multiple files with the same definitions and build constraints to compile only one at a time
		if _, ok := o.obfuscatedNames[fullName(obj)]; ok {
			continue
		}
		o.obfuscatedNames[fullName(obj)] = o.nextName(obj.Exported())
	}
}

func fullName(obj types.Object) string {
	if !obj.Exported() {
		return obj.Id()
	} else {
		return obj.Pkg().Name() + "." + obj.Id()
	}
}

func (o *Obfuscator) obfuscateAST() {
	for ident, obj := range o.info.Defs {
		if obj == nil || obj.Pkg() == nil {
			continue
		}
		if _, ok := obj.(*types.PkgName); ok {
			continue
		}
		if newName, ok := o.obfuscatedNames[fullName(obj)]; ok {
			ident.Name = newName
		}
	}
	for ident, obj := range o.info.Uses {
		if obj == nil || obj.Pkg() == nil {
			continue
		}
		if _, ok := obj.(*types.PkgName); ok {
			continue
		}
		if newName, ok := o.obfuscatedNames[fullName(obj)]; ok {
			ident.Name = newName
		}
	}
}

func (o *Obfuscator) writeASTFile(filename string, file *ast.File) error {
	dirname := path.Dir(filename)
	err := os.MkdirAll(dirname, fs.ModePerm)
	if err != nil {
		return errors.WithMessagef(err, "Unable to create dir %s", dirname)
	}

	outFile, err := os.Create(filename)
	if err != nil {
		return errors.WithMessagef(err, "Unable to create file %s", filename)
	}
	defer func() { _ = outFile.Close() }()

	err = printer.Fprint(outFile, o.fset, file)
	if err != nil {
		return errors.WithMessagef(err, "Unable to write AST to file %s", filename)
	}
	return nil
}

func (o *Obfuscator) writeAST() error {
	for _, astPkg := range o.astPkgs {
		for filename, file := range astPkg.Files {
			filename = strings.Replace(filename, o.sourcePath, o.targetPath, 1)
			err := o.writeASTFile(filename, file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *Obfuscator) nextName(isExported bool) string {
	o.currentName++
	if isExported {
		return "A" + strconv.Itoa(o.currentName)
	}
	return "a" + strconv.Itoa(o.currentName)
}
