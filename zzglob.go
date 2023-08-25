// Package zzglob implements a file path walker.
package zzglob

import (
	"io/fs"
)

// Glob globs for files matching the pattern in a filesystem.
func Glob(fsys fs.FS, pattern string, f fs.WalkDirFunc, traverseSymlinks bool) error {
	root, start, err := parse(pattern)
	if err != nil {
		return err
	}

	// This is a queue of roots, and the partial matching states to get to that
	// root.
	type globState struct {
		root   string
		states map[*state]struct{}
	}
	queue := []globState{{root, singleton(start)}}

	for len(queue) > 0 {
		gs := queue[0]
		queue = queue[1:]

		if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
			// Match some of the path against the pattern.
			states := matchSegment(gs.states, path)

			// Did it match in any way? The caller eagerly wants to know.
			if len(states) == 0 {
				if d.IsDir() {
					// Skip - not interested in anything in this directory.
					return fs.SkipDir
				}

				// This non-directory thing doesn't match. Don't return
				// fs.SkipDir, since that skips the remainder of the directory.
				return nil
			}

			if err != nil {
				// Report the error to the callback.
				return f(path, d, err)
			}

			// So we matched, either partially or fully.
			terminal := false
			for s := range states {
				if s.terminal() {
					terminal = true
					break
				}
			}

			// Did the pattern match completely?
			if terminal {
				// Give it to the callback.
				return f(path, d, err)
			}

			// The pattern matched only partially.
			// Are we traversing symlinks? Is it a symlink?
			if traverseSymlinks && d.Type()&fs.ModeSymlink != 0 {
				// Walk the symlink by enqueueing a new root.
				queue = append(queue, globState{
					root:   path,
					states: states,
				})
			}

			// Continue walking as normal.
			return nil

		}); err != nil {
			return err
		}
	}
	return nil
}
