// Package zzglob implements a file path walker.
package zzglob

import (
	"io/fs"
	"os"
	"path"
	"strings"
)

// Glob globs for files matching the pattern in a filesystem.
func (p *Pattern) Glob(f fs.WalkDirFunc, traverseSymlinks bool) error {
	if p.initial == nil {
		fi, err := os.Stat(p.root)
		if err != nil {
			return f(p.root, fs.FileInfoToDirEntry(fi), err)
		}
		return f(p.root, nil, nil)
	}

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

		//println("starting walk at", gs.root, "with", len(gs.states), "states")

		if err := fs.WalkDir(os.DirFS(gs.root), ".", func(fp string, d fs.DirEntry, err error) error {
			// Yes yes, very good.
			if fp == "." {
				return nil
			}

			// Directories have a trailing slash for matching.
			// (Symlinks to other directories won't get a slash here.)
			if d != nil && d.IsDir() && !strings.HasSuffix(fp, "/") {
				fp += "/"
			}

			// Match p against the state machine.
			states := matchSegment(gs.states, fp)

			// Some debugging code that might be handy:
			// println("matchSegment(", len(gs.states), fp, ") ->", len(states), "states")
			// if len(states) > 0 {
			// 	p.WriteDot(os.Stderr, states)
			// }

			// Join fp back onto root, for passing back to f and for other
			// operations.
			fp = path.Join(gs.root, fp)

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
				return f(fp, d, err)
			}

			// So we matched, either partially or fully.
			terminal := false
			for s := range states {
				if s.Terminal {
					terminal = true
					break
				}
			}

			// Did the pattern match completely?
			if terminal {
				// Give it to the callback.
				return f(fp, d, err)
			}

			// The pattern matched only partially.
			// Are we traversing symlinks? Is it a symlink?
			if traverseSymlinks {
				fi, err := os.Lstat(fp)
				if err != nil {
					// Can't lstat, can't tell if link.
					return f(fp, fs.FileInfoToDirEntry(fi), err)
				}
				if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
					// Not a symlink.
					return nil
				}

				// Make sure we're matching a directory here...
				states = matchSegment(states, "/")
				if len(states) == 0 {
					return nil
				}

				// Walk the symlink by enqueueing a new root.
				// fs.WalkDir doesn't walk symlinks unless it is the root path!
				queue = append(queue, globState{
					root:   fp,
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
