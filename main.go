package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		Source string   `short:"s" long:"source" description:"Directory with code to obfuscate" required:"true"`
		Target string   `short:"t" long:"target" description:"Directory to write obfuscated code to" required:"true"`
		Ignore []string `short:"i" long:"ignore" description:"Directory to ignore"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	opts.Source, err = filepath.Abs(opts.Source)
	if err != nil {
		log.Fatalf("Unable to get absolute source path %s: %v", opts.Source, err)
	}
	opts.Target, err = filepath.Abs(opts.Target)
	if err != nil {
		log.Fatalf("Unable to get absolute target path %s: %v", opts.Target, err)
	}

	for i, ignore := range opts.Ignore {
		opts.Ignore[i], err = filepath.Abs(ignore)
		if err != nil {
			log.Fatalf("Unable to get absolute ignore path %s: %v", ignore, err)
		}
	}

	if _, err := os.Stat(opts.Target); os.IsNotExist(err) {
		dir, err := os.Open(opts.Target)
		if err == nil {
			_, err = dir.Readdirnames(1)
			if err != io.EOF {
				log.Fatalf("Target directory %s is not empty!", opts.Target)
			}
		}
	}

	err = os.MkdirAll(opts.Target, fs.ModePerm)
	if err != nil {
		log.Fatalf("Unable to create target directory %s: %v", opts.Target, err)
	}

	obfuscator, err := NewObfuscator(opts.Source, opts.Target, opts.Ignore)
	if err != nil {
		log.Fatal(err)
	}
	err = obfuscator.Obfuscate()
	if err != nil {
		log.Fatal(err)
	}
}
