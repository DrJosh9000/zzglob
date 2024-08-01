// The zzglob command searches for files with paths matching a pattern.
//
// Example:
//
//	$ zzglob '**/*_test.go'
//	fixtures/spec/cmd/cmd_test.go
//	fixtures/spec/foo_test.go
//	glob_test.go
//	match_test.go
//	multiglob_test.go
//	pattern_test.go
//	tokeniser_test.go
package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/DrJosh9000/zzglob"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s pattern\n", os.Args[0])
		os.Exit(1)
	}

	p, err := zzglob.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse pattern %q: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	err = p.Glob(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error at path %q: %v\n", path, err)
			return nil
		}
		fmt.Println(path)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't perform file system walk: %v\n", err)
	}
}
