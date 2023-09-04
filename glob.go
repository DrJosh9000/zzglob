// Package zzglob implements a file path walker.
package zzglob

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"strings"
)

// Glob globs for files matching the pattern in a filesystem.
func (p *Pattern) Glob(f fs.WalkDirFunc, opts ...GlobOption) error {
	if f == nil {
		return errors.New("nil WalkDirFunc in arg to Glob")
	}

	if p.initial == nil {
		fi, err := os.Stat(p.root)
		if err != nil {
			return f(p.root, fs.FileInfoToDirEntry(fi), err)
		}
		return f(p.root, nil, nil)
	}

	gs := globState{
		opts: &globConfig{
			TraverseSymlinks: true,
			Callback:         f,
		},
		root:   p.root,
		states: singleton(p.initial),
	}

	for _, o := range opts {
		o(gs.opts)
	}

	//println("starting walk at", gs.root, "with", len(gs.states), "states")
	return fs.WalkDir(os.DirFS(gs.root), ".", gs.walkDirFunc)
}

// GlobOption is the type for options that apply to Glob.
type GlobOption = func(*globConfig)

type globConfig struct {
	TraverseSymlinks bool
	Callback         fs.WalkDirFunc
}

// TraverseSymlinks enables or disables the traversal of symlinks during
// globbing. It is enabled by default.
func TraverseSymlinks(traverse bool) GlobOption {
	return func(opts *globConfig) {
		opts.TraverseSymlinks = traverse
	}
}

type globState struct {
	opts   *globConfig
	root   string
	states map[*state]struct{}
}

func (gs *globState) walkDirFunc(fp string, d fs.DirEntry, err error) error {
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
		return gs.opts.Callback(fp, d, err)
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
		return gs.opts.Callback(fp, d, err)
	}

	// The pattern matched only partially.
	// Are we traversing symlinks? Is it a symlink?
	if !gs.opts.TraverseSymlinks {
		return nil
	}

	// It's all symlink from this point.

	fi, err := os.Lstat(fp)
	if err != nil {
		// Can't lstat, can't tell if link.
		return gs.opts.Callback(fp, fs.FileInfoToDirEntry(fi), err)
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

	// Walk the symlink by... recursion.
	// fs.WalkDir doesn't walk symlinks unless it is the root path... in
	// which case it does!
	next := globState{
		opts:   gs.opts,
		root:   fp,
		states: states,
	}
	//println("starting walk at", next.root, "with", len(next.states), "states")
	return fs.WalkDir(os.DirFS(next.root), ".", next.walkDirFunc)
}
