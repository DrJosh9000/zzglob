// Package zzglob implements a file path walker.
package zzglob

import (
	"errors"
	"io/fs"
	"os"
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
		cfg: &globConfig{
			TraverseSymlinks: true,
			Callback:         f,
			FS:               os.DirFS("."),
		},
		root:   p.root,
		states: singleton(p.initial),
	}

	for _, o := range opts {
		o(gs.cfg)
	}

	if gs.cfg.FS == nil {
		return errors.New("nil filesystem in options to Glob")
	}

	//println("starting walk at", gs.root, "with", len(gs.states), "states")
	return fs.WalkDir(gs.cfg.FS, gs.root, gs.walkDirFunc)
}

// GlobOption functions optionally alter how Glob operates.
type GlobOption = func(*globConfig)

type globConfig struct {
	TraverseSymlinks bool
	Callback         fs.WalkDirFunc
	FS               fs.FS
}

// WithFilesystem allows overriding the default filesystem (os.DirFS(".")).
func WithFilesystem(fs fs.FS) GlobOption {
	return func(cfg *globConfig) {
		cfg.FS = fs
	}
}

// TraverseSymlinks enables or disables the traversal of symlinks during
// globbing. It is enabled by default.
func TraverseSymlinks(traverse bool) GlobOption {
	return func(cfg *globConfig) {
		cfg.TraverseSymlinks = traverse
	}
}

type globState struct {
	cfg    *globConfig
	root   string
	states map[*state]struct{}
}

func (gs *globState) walkDirFunc(fp string, d fs.DirEntry, err error) error {
	if fp == gs.root {
		// Yes...?
		return nil
	}

	// Directories have a trailing slash for matching.
	// (Symlinks to other directories won't get a slash here.)

	trimmed := strings.TrimPrefix(fp, gs.root)

	// println("trimmed =", trimmed)

	// Match p against the state machine.
	states := matchSegment(gs.states, trimmed)

	// If it's a directory the pattern should match another /
	if d != nil && d.IsDir() && !strings.HasSuffix(fp, "/") {
		states = matchSegment(states, "/")
	}

	// Some debugging code that might be handy:
	//
	// println("matchSegment(", len(gs.states), trimmed, ") ->", len(states), "states")
	// if len(states) > 0 {
	// 	p.WriteDot(os.Stderr, states)
	// }

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
		return gs.cfg.Callback(fp, d, err)
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
		return gs.cfg.Callback(fp, d, err)
	}

	// The pattern matched only partially.
	// Are we traversing symlinks?
	if !gs.cfg.TraverseSymlinks {
		// Nope - just keep walking.
		return nil
	}

	// It's all symlink handling from this point.
	if d != nil && d.Type()&fs.ModeSymlink != fs.ModeSymlink {
		// Not a symlink.
		return nil
	}

	// Walk the symlink by... recursion.
	// fs.WalkDir doesn't walk symlinks unless it is the root path... in
	// which case it does!
	next := globState{
		cfg:    gs.cfg,
		root:   fp,
		states: states,
	}
	//println("walking symlink at", next.root, "with", len(next.states), "states")
	return fs.WalkDir(gs.cfg.FS, next.root, next.walkDirFunc)
}
