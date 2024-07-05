package main

import (
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Src    string `short:"s" long:"source" description:"Directory with code to obfuscate" required:"true"`
	Target string `short:"t" long:"target" description:"Directory to write obfuscated code to" required:"true"`
}

func main() {
	// Parse args
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}
}
