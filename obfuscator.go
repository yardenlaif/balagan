package main

import (
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/packages"
)

// Don't obfuscate these names as they hold semantic meaning
var exclude = map[string]struct{}{"main": {}, "init": {}, "_": {}}

type Obfuscator struct {
	interfaces      map[*types.Interface]struct{}
	info            *types.Info
	currentName     int
	obfuscatedNames map[string]string
	astFiles        []*ast.File
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
		info: &types.Info{
			Defs:   make(map[*ast.Ident]types.Object),
			Types:  make(map[ast.Expr]types.TypeAndValue),
			Uses:   make(map[*ast.Ident]types.Object),
			Scopes: make(map[ast.Node]*types.Scope),
		},
	}

	// This is necessary for the importer to work
	err := os.Chdir(sourcePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "Unable to change directory to source directory %s", sourcePath)
	}

	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedDeps | packages.NeedTypesInfo | packages.NeedImports | packages.NeedSyntax, Fset: o.fset}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}

	typesPkgs := make(map[*types.Package]struct{})
	for _, pkg := range pkgs {
		typesPkgs[pkg.Types] = struct{}{}
		o.astFiles = append(o.astFiles, pkg.Syntax...)
		maps.Copy(o.info.Types, pkg.TypesInfo.Types)
		maps.Copy(o.info.Defs, pkg.TypesInfo.Defs)
		maps.Copy(o.info.Uses, pkg.TypesInfo.Uses)
		maps.Copy(o.info.Scopes, pkg.TypesInfo.Scopes)
	}
	o.interfaces = findInterfaces(maps.Keys(typesPkgs))

	return o, nil
}

func (o *Obfuscator) Obfuscate() error {
	o.createObfuscatedNames()
	o.obfuscateAST()
	removeComments(o.astFiles)
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
	for _, file := range o.astFiles {
		filename := strings.Replace(o.fset.Position(file.Package).Filename, o.sourcePath, o.targetPath, 1)
		err := o.writeASTFile(filename, file)
		if err != nil {
			return err
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
