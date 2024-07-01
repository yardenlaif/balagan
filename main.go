package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"golang.org/x/exp/maps"
)

var opts struct {
	Src    string `short:"s" long:"source" description:"Directory with code to obfuscate" required:"true"`
	Target string `short:"t" long:"target" description:"Directory to write obfuscated code to" required:"true"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}
	opts.Src, _ = filepath.Abs(opts.Src)
	opts.Target, _ = filepath.Abs(opts.Target)

	os.Chdir(opts.Src)
	os.MkdirAll(opts.Target, 0777)
	fset := token.NewFileSet()
	astPkgs := make(map[*ast.Package]struct{})
	conf := &types.Config{Importer: importer.ForCompiler(fset, "source", nil), Error: func(err error) {}}
	info := &types.Info{
		Scopes: make(map[ast.Node]*types.Scope),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
	}

	typesPkgs := make(map[*types.Package]struct{})
	filepath.WalkDir(opts.Src, func(path string, e fs.DirEntry, _ error) error {
		if e.Type() != fs.ModeDir {
			if !strings.HasSuffix(e.Name(), ".go") {
				src, _ := filepath.Abs(e.Name())
				dest := strings.Replace(src, opts.Src, opts.Target, 1)
				cmd := exec.Command("cp", src, dest)
				cmd.Run()
			}
			return nil
		}
		pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("Unable to parse source directory, error:\n%v", err)
		}
		for _, astPkg := range pkgs {
			astPkgs[astPkg] = struct{}{}
			pkg, _ := conf.Check(path, fset, maps.Values(astPkg.Files), info)
			typesPkgs[pkg] = struct{}{}
		}
		return nil
	})

	obfuscate(opts.Src, opts.Target, fset, maps.Keys(astPkgs), maps.Keys(typesPkgs), info)

	// NewPackageObfuscator(astPkgs, interfaces)
}
