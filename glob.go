// Package zzglob implements a file path walker.
package zzglob

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const globSymlinkRecursionLimit = 1000

// Glob globs for files matching the pattern in a filesystem.
func (p *Pattern) Glob(f fs.WalkDirFunc, opts ...GlobOption) error {
	if f == nil {
		return errors.New("nil WalkDirFunc in arg to Glob")
	}

	cfg := &globConfig{
		translateSlashes: true,
		traverseSymlinks: true,
		callback:         f,
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		o(cfg)
	}

	// p.root always uses forward slashes. Translate (if needed)?
	cleanRoot := path.Clean(p.root)
	osRoot := cleanRoot
	if cfg.translateSlashes {
		osRoot = filepath.FromSlash(cleanRoot)
	}

	if p.initial == nil {
		if cfg.filesystem == nil {
			// The fastest way to stat the file is... to stat the file.
			fi, err := os.Stat(osRoot)
			if err := f(osRoot, fs.FileInfoToDirEntry(fi), err); err != nil {
				if errors.Is(err, fs.SkipDir) || errors.Is(err, fs.SkipAll) {
					return nil
				}
				return err
			}
		} else {
			// Assume root sits at that path within the provided fs.FS.
			fi, err := fs.Stat(cfg.filesystem, cleanRoot)
			if err := f(osRoot, fs.FileInfoToDirEntry(fi), err); err != nil {
				if errors.Is(err, fs.SkipDir) || errors.Is(err, fs.SkipAll) {
					return nil
				}
				return err
			}
		}
		return nil
	}

	gs := globState{
		cfg:    cfg,
		root:   cleanRoot,
		fs:     cfg.filesystem,
		states: singleton(p.initial),
	}

	// Filesystem override?
	if gs.fs == nil {
		// Wasn't overridden
		gs.fs = os.DirFS(osRoot)
	} else {
		subfs, err := fs.Sub(cfg.filesystem, cleanRoot)
		if err != nil {
			// That's unfortunate.
			return fmt.Errorf("pattern root %q not valid within provided filesystem: %w", cleanRoot, err)
		}
		gs.fs = subfs
	}

	gs.logf("starting walk in fsys %v, root %q at . with %d states\n", gs.fs, gs.root, len(gs.states))
	return fs.WalkDir(gs.fs, ".", gs.walkDirFunc)
}

type globState struct {
	depth  int
	cfg    *globConfig
	root   string
	fs     fs.FS
	states stateSet
}

func (gs *globState) logf(f string, v ...any) {
	if gs.cfg.traceLogger != nil {
		fmt.Fprintf(gs.cfg.traceLogger, f, v...)
	}
}

func (gs *globState) walkDirFunc(fp string, d fs.DirEntry, err error) error {
	gs.logf("globState.walkDirFunc(%q, %v, %v)\n", fp, d, err)

	if gs.depth > globSymlinkRecursionLimit {
		return fmt.Errorf("recursion limit %d reached; possible symlink cycle", globSymlinkRecursionLimit)
	}

	if fp == "." {
		gs.logf("fast path for .\n")
		if gs.cfg.walkIntermediateDirs {
			full := gs.root
			if gs.cfg.translateSlashes {
				full = filepath.FromSlash(full)
			}
			return gs.cfg.callback(full, d, err)
		}
		return nil
	}

	// Directories have a trailing slash for matching.
	// (Symlinks to other directories won't get a slash here.)

	// Rage (match fp) against the (state) machine.
	states := matchSegment(gs.states, fp)

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

	if terminal || err != nil || (gs.cfg.walkIntermediateDirs && d.IsDir()) {
		gs.logf("fully matched, error, or intermediate dir! calling callback\n")
		if gs.cfg.translateSlashes {
			full = filepath.FromSlash(full)
		}
		return gs.cfg.callback(full, d, err)
	}

	// The pattern matched only partially...
	// Are we traversing symlinks?
	if !gs.cfg.traverseSymlinks {
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

	// Because we only traverse symlinks to directories, the pattern must match
	// another /.
	states = matchSegment(states, "/")
	if len(states) == 0 {
		gs.logf("pattern did not match additional /; continuing walk\n")
		return nil
	}

	subfs, err := fs.Sub(gs.fs, fp)
	if err != nil {
		gs.logf("error from fs.Sub(gs.fsys, %q): %v - passing to callback\n", fp, err)
		return gs.cfg.callback(fp, d, err)
	}

	// Walk the symlink by... recursion.
	// fs.WalkDir doesn't walk symlinks unless it is the root path... in
	// which case it does!
	next := globState{
		depth:  gs.depth + 1,
		cfg:    gs.cfg,
		root:   full,
		fs:     subfs,
		states: states,
	}

	gs.logf("starting symlink walk in fsys %v, root %q at . with %d states\n", subfs, next.root, len(gs.states))
	return fs.WalkDir(subfs, ".", next.walkDirFunc)
}
