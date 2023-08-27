// Package zzglob implements a file path walker.
package zzglob

import (
	"io/fs"
	"os"
	"path"
)

// Glob globs for files matching the pattern in a filesystem.
func (p *Pattern) Glob(f fs.WalkDirFunc, traverseSymlinks bool) error {
	// This is a queue of roots, and the partial matching states to get to that
	// root.
	type globState struct {
		root   string
		states map[*state]struct{}
	}
	queue := []globState{{p.root, singleton(p.initial)}}

	for len(queue) > 0 {
		gs := queue[0]
		queue = queue[1:]

		if err := fs.WalkDir(os.DirFS(gs.root), ".", func(p string, d fs.DirEntry, err error) error {
			// Yes yes, very good.
			if p == "." {
				return nil
			}

			// Match some of the path against the pattern.
			states := matchSegment(gs.states, p)

			// Join it back onto root, for passing back to f and for other
			// operations.
			p = path.Join(gs.root, p)

			// Did it match in any way?
			if len(states) == 0 {
				if d != nil && d.IsDir() {
					// Skip - not interested in anything in this directory.
					return fs.SkipDir
				}

				// This non-directory thing doesn't match. Don't return
				// fs.SkipDir, since that skips the remainder of the directory.
				return nil
			}

			if err != nil {
				// Report the error to the callback.
				return f(p, d, err)
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
				return f(p, d, err)
			}

			// The pattern matched only partially.
			// Are we traversing symlinks? Is it a symlink?
			if traverseSymlinks && d.Type()&fs.ModeSymlink != 0 {
				// There is no fs.ReadLinkFS, therefore we need to use os...
				// https://github.com/golang/go/issues/49580
				target, err := os.Readlink(p)
				if err != nil {
					// Report the error to the callback.
					return f(p, d, err)
				}

				// Walk the symlink by enqueueing a new root.
				queue = append(queue, globState{
					root:   target,
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
