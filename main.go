package main

import (
	"go/parser"
	"go/token"

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

	fset := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fset, opts.Src, nil, 0)
	if err != nil {
		// TODO: Deal with this differently
		panic(err)
	}

	obfuscate(opts.Src, fset, maps.Values(astPkgs))

	// NewPackageObfuscator(astPkgs, interfaces)
}
