// Package zzglob implements a file path walker.
package zzglob

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Glob globs for files matching the pattern in a filesystem.
func (p *Pattern) Glob(f fs.WalkDirFunc, opts ...GlobOption) error {
	if f == nil {
		return errors.New("nil WalkDirFunc in arg to Glob")
	}

	gs := globState{
		cfg: &globConfig{
			TraverseSymlinks: true,
			TraceLogs:        nil,
			Callback:         f,
			Filesystem:       os.DirFS(p.root),
		},
		root:   p.root,
		states: singleton(p.initial),
	}

	for _, o := range opts {
		o(gs.cfg)
	}

	// Filesystem override?
	if gs.cfg.Filesystem == nil {
		return errors.New("nil filesystem in options to Glob")
	}
	gs.fsys = gs.cfg.Filesystem

	if gs.cfg.TranslateSlashes {
		gs.root = filepath.ToSlash(p.root)
	}

	if p.initial == nil {
		fi, err := fs.Stat(gs.fsys, p.root)
		return f(p.root, fs.FileInfoToDirEntry(fi), err)
	}

	gs.logf("starting walk in fsys %v, root %q at . with %d states\n", gs.fsys, gs.root, len(gs.states))
	return fs.WalkDir(gs.fsys, ".", gs.walkDirFunc)
}

// GlobOption functions optionally alter how Glob operates.
type GlobOption = func(*globConfig)

type globConfig struct {
	TraverseSymlinks bool
	TranslateSlashes bool
	TraceLogs        io.Writer
	Callback         fs.WalkDirFunc
	Filesystem       fs.FS
}

// WithFilesystem allows overriding the default filesystem (os.DirFS(".")).
func WithFilesystem(fs fs.FS) GlobOption {
	return func(cfg *globConfig) {
		cfg.Filesystem = fs
	}
}

// WithTraceLogs logs debugging information for debugging Glob itself to the
// provided writer. Disabled by default.
func WithTraceLogs(out io.Writer) GlobOption {
	return func(cfg *globConfig) {
		cfg.TraceLogs = out
	}
}

// TraverseSymlinks enables or disables the traversal of symlinks during
// globbing. It is enabled by default.
func TraverseSymlinks(traverse bool) GlobOption {
	return func(cfg *globConfig) {
		cfg.TraverseSymlinks = traverse
	}
}

// TranslateSlashes enables or disables translating to and from fs.FS paths
// (always with forward slashes, / ) using filepath.FromSlash. This applies to
// both the matching pattern and filepaths passed to the callback, and is
// typically required on Windows. Enabled by default.
func TranslateSlashes(enable bool) GlobOption {
	return func(cfg *globConfig) {
		cfg.TranslateSlashes = enable
	}
}

type globState struct {
	cfg    *globConfig
	fsys   fs.FS
	root   string
	states map[*state]struct{}
}

func (gs *globState) logf(f string, v ...any) {
	if gs.cfg.TraceLogs != nil {
		fmt.Fprintf(gs.cfg.TraceLogs, f, v...)
	}
}

func (gs *globState) walkDirFunc(fp string, d fs.DirEntry, err error) error {
	gs.logf("globState.walkDirFunc(%q, %v, %v)\n", fp, d, err)
	if fp == "." {
		gs.logf("fast path for .\n")
		return nil
	}

	// Directories have a trailing slash for matching.
	// (Symlinks to other directories won't get a slash here.)

	// Rage (match /fp) against the (state) machine.
	states := matchSegment(gs.states, "/"+fp)

	// If it's a directory the pattern should match another /
	if d != nil && d.IsDir() && !strings.HasSuffix(fp, "/") {
		states = matchSegment(states, "/")
	}

	gs.logf("matchSegment(%d states, %q) -> %d states\n", len(gs.states), fp, len(states))

	terminal := false
	for s := range states {
		if s.Terminal {
			terminal = true
			gs.logf("\t(at least one terminal state)\n")
			break
		}
	}

	// Did it match in any way?
	if len(states) == 0 {
		if d != nil && d.IsDir() {
			// Skip - not interested in anything in this directory.
			gs.logf("directory didn't match at all; returning fs.SkipDir\n")
			return fs.SkipDir
		}

		// This non-directory thing doesn't match. Don't return
		// fs.SkipDir, since that skips the remainder of the directory.
		gs.logf("non-directory didn't match at all; returning nil\n")
		return nil
	}

	full := path.Join(gs.root, fp)
	gs.logf("full = %q\n", full)

	if terminal || err != nil {
		gs.logf("fully matched, or error! calling callback\n")
		if gs.cfg.TranslateSlashes {
			full = filepath.FromSlash(full)
		}
		return gs.cfg.Callback(full, d, err)
	}

	// The pattern matched only partially...
	// Are we traversing symlinks?
	if !gs.cfg.TraverseSymlinks {
		// Nope - just keep walking.
		gs.logf("symlink traversal disabled; continuing walk\n")
		return nil
	}

	// It's all symlink handling from this point.
	if d == nil || d.Type()&fs.ModeSymlink == 0 {
		// Not a symlink.
		gs.logf("not a symlink; continuing walk\n")
		return nil
	}

	subfs, err := fs.Sub(gs.fsys, fp)
	if err != nil {
		gs.logf("error from fs.Sub(gs.fsys, %q): %v\n", fp, err)
		return err
	}

	// Walk the symlink by... recursion.
	// fs.WalkDir doesn't walk symlinks unless it is the root path... in
	// which case it does!
	next := globState{
		cfg:    gs.cfg,
		fsys:   subfs,
		root:   full,
		states: states,
	}

	gs.logf("starting symlink walk in fsys %v, root %q at . with %d states\n", next.fsys, next.root, len(gs.states))
	return fs.WalkDir(next.fsys, ".", next.walkDirFunc)
}
