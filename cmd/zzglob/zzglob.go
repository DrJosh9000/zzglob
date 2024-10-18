// The zzglob command searches for files with paths matching a pattern.
//
// Example:
//
//	$ zzglob -include '**/*_test.go'
//	fixtures/spec/cmd/cmd_test.go
//	fixtures/spec/foo_test.go
//	glob_test.go
//	match_test.go
//	multiglob_test.go
//	pattern_test.go
//	tokeniser_test.go
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"time"

	"drjosh.dev/zzglob"
)

func main() {
	includePattern := flag.String("include", "", "Glob pattern for matching files to include")
	excludePattern := flag.String("exclude", "", "Glob pattern for matching files or directories to exclude")
	listing := flag.Bool("l", false, "If enabled, extra file information is printed for each match")
	enableTrace := flag.Bool("trace", false, "If enabled, tracing information is logged to stderr")
	flag.Parse()

	incl, err := zzglob.Parse(*includePattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse pattern %q: %v\n", *includePattern, err)
		os.Exit(1)
	}
	if *enableTrace {
		incl.WriteDot(os.Stderr, nil)
	}

	var opts []zzglob.GlobOption
	if *enableTrace {
		opts = append(opts, zzglob.WithTraceLogs(os.Stderr))
	}
	var excl *zzglob.Pattern
	if *excludePattern != "" {
		ep, err := zzglob.Parse(*excludePattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't parse exclude pattern %q: %v\n", *excludePattern, err)
			os.Exit(1)
		}
		excl = ep
		opts = append(opts, zzglob.WalkIntermediateDirs(true))
		if *enableTrace {
			excl.WriteDot(os.Stderr, nil)
		}
	}

	err = incl.Glob(
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error at path %q: %v\n", path, err)
				return nil
			}

			if excl != nil {
				if excl.Match(path) {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
			}

			if d.IsDir() {
				return nil
			}

			if *listing {
				fi, err := d.Info()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error at path %q: %v\n", path, err)
					return nil
				}
				fmt.Printf("%v\t%d\t%s\t%s\n", fi.Mode(), fi.Size(), fi.ModTime().Format(time.RFC3339), path)
			} else {
				fmt.Println(path)
			}

			return nil
		},
		opts...,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't perform file system walk: %v\n", err)
	}
}
