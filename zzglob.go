// Package zzglob implements a file path walker.
package zzglob

import (
	"io/fs"
)

// Glob globs for files matching the pattern in a filesystem.
func Glob(fsys fs.FS, pattern string, f fs.WalkDirFunc) error {
	root, start, err := parse(pattern)
	if err != nil {
		return err
	}

	// New roots are added in order to traverse symlinks.
	type globState struct {
		root   string
		states map[*state]struct{}
	}
	queue := []globState{{root, singleton(start)}}
	for len(queue) > 0 {
		gs := queue[0]
		queue = queue[1:]

		if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
			out := matchSegment(gs.states, path)
			if len(out) == 0 && d.IsDir() {
				// Skip - not interested in anything in this directory.
				return fs.SkipDir
			}
			if err != nil {
				return f(path, d, err)
			}

			return nil

		}); err != nil {
			return err
		}
	}
	return nil
}
