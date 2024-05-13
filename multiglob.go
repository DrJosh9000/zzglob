package zzglob

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// MultiGlob is like [Pattern.Glob], but globs multiple patterns simultaneously.
// The main idea is to group the patterns by root to avoid multiple different
// calls to [fs.WalkDir] (reducing filesystem I/O), and then to use [fs.WalkDir]
// on each root in parallel.
// As a result, files can be walked globbed multiple times, but only if distinct
// overlapping roots appear in different input patterns.
// You should either make sure that the callback f is safe to call concurrently
// from multiple goroutines, or set GoroutineLimit to 1.
func MultiGlob(ctx context.Context, patterns []*Pattern, f fs.WalkDirFunc, opts ...GlobOption) error {
	if f == nil {
		return errors.New("nil WalkDirFunc in arg to MultiGlob")
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

	// Group patterns by cleaned root
	byRoot := make(map[string][]*Pattern)
	for _, p := range patterns {
		cleanRoot := path.Clean(p.root)
		byRoot[cleanRoot] = append(byRoot[cleanRoot], p)
	}

	// Spin up this many worker goroutines.
	if cfg.goroutines <= 0 || cfg.goroutines > len(byRoot) {
		cfg.goroutines = len(byRoot)
	}
	workCh := make(chan multiglobWork)
	wctx, cancel := context.WithCancelCause(ctx)
	var wg sync.WaitGroup
	for i := 0; i < cfg.goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := multiglobWorker(ctx, cfg, workCh); err != nil {
				cancel(err)
			}
		}()
	}

	// Feed work to the workers
	for root, patts := range byRoot {
		work := multiglobWork{
			root:     root,
			patterns: patts,
		}
		select {
		case <-wctx.Done():
			return context.Cause(wctx)

		case workCh <- work:
			// work has been fed
		}
	}
	close(workCh)

	wg.Wait()
	return context.Cause(wctx)
}

type multiglobWork struct {
	root     string
	patterns []*Pattern
}

func multiglobWorker(ctx context.Context, cfg *globConfig, workCh <-chan multiglobWork) error {
	for {
		var root string
		var patterns []*Pattern

		select {
		case work, open := <-workCh:
			if !open {
				return nil
			}
			root, patterns = work.root, work.patterns

		case <-ctx.Done():
			return ctx.Err()

		}

		// root always uses forward slashes. Translate (if needed)?
		osRoot := root
		if cfg.translateSlashes {
			osRoot = filepath.FromSlash(root)
		}

		// Accumulate all the initial states for the patterns in the group.
		// Invoke the callback for any patterns that are fully specified.
		states := make(map[*state]struct{})
		for _, p := range patterns {
			if p.initial == nil {
				if cfg.filesystem == nil {
					// The fastest way to stat the file is... to stat the file.
					fi, err := os.Stat(osRoot)
					if err := cfg.callback(osRoot, fs.FileInfoToDirEntry(fi), err); err != nil {
						if errors.Is(err, fs.SkipDir) || errors.Is(err, fs.SkipAll) {
							return nil
						}
						return err
					}
				} else {
					// Assume root sits at that path within the provided [fs.FS].
					fi, err := fs.Stat(cfg.filesystem, root)
					if err := cfg.callback(osRoot, fs.FileInfoToDirEntry(fi), err); err != nil {
						if errors.Is(err, fs.SkipDir) || errors.Is(err, fs.SkipAll) {
							return nil
						}
						return err
					}
				}
				continue
			}
			states[p.initial] = struct{}{}
		}

		gs := globState{
			cfg:    cfg,
			root:   root,
			fs:     cfg.filesystem,
			states: states,
		}

		// Filesystem override?
		if gs.fs == nil {
			// Wasn't overridden
			gs.fs = os.DirFS(osRoot)
		} else {
			subfs, err := fs.Sub(gs.fs, root)
			if err != nil {
				// That's unfortunate.
				return fmt.Errorf("pattern root %q not valid within provided filesystem: %w", root, err)
			}
			gs.fs = subfs
		}

		gs.logf("starting walk in fsys %v, root %q at . with %d states\n", gs.fs, root, len(gs.states))
		if err := fs.WalkDir(gs.fs, ".", func(path string, d fs.DirEntry, err error) error {
			// Check that work isn't cancelled yet
			if err := ctx.Err(); err != nil {
				return err
			}
			return gs.walkDirFunc(path, d, err)
		}); err != nil {
			return err
		}
	}
}
