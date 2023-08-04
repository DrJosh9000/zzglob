// Package zzglob implements a file path walker.
package zzglob

import (
	"io/fs"
	"path/filepath"
)

// Glob globs for files matching the pattern.
func Glob(pattern string, f fs.WalkDirFunc) error {
	root, _, err := parse(pattern)
	if err != nil {
		return err
	}

	// Roots holds roots to walk.
	// New roots are added in order to traverse symlinks.
	roots := []string{root}
	for len(roots) > 0 {
		root := roots[0]
		roots = roots[1:]

		if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// if pattern.match(path) {
			// 	if err := f(path, d, err); err != nil {
			// 		return err
			// 	}
			// }
			// if pattern.partialMatch(path) {
			// 	// something like that
			// }

			return nil

		}); err != nil {
			return err
		}
	}
	return nil
}
