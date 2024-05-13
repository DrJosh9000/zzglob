package zzglob

import (
	"io"
	"io/fs"
)

// GlobOption functions optionally alter how Glob operates.
type GlobOption = func(*globConfig)

type globConfig struct {
	traverseSymlinks     bool
	translateSlashes     bool
	walkIntermediateDirs bool
	traceLogger          io.Writer
	filesystem           fs.FS
	goroutines           int // only used by MultiGlob

	callback fs.WalkDirFunc // the required arg to Glob
}

// WithFilesystem allows overriding the default filesystem. By default os.DirFS
// is used to wrap file/directory access in an `fs.FS`.
func WithFilesystem(fs fs.FS) GlobOption {
	return func(cfg *globConfig) {
		cfg.filesystem = fs
	}
}

// WithTraceLogs logs debugging information for debugging Glob itself to the
// provided writer. Disabled by default.
func WithTraceLogs(out io.Writer) GlobOption {
	return func(cfg *globConfig) {
		cfg.traceLogger = out
	}
}

// TraverseSymlinks enables or disables the traversal of symlinks during
// globbing. It is enabled by default.
func TraverseSymlinks(traverse bool) GlobOption {
	return func(cfg *globConfig) {
		cfg.traverseSymlinks = traverse
	}
}

// TranslateSlashes enables or disables translating to and from fs.FS paths
// (always with forward slashes, / ) using filepath.FromSlash. This applies to
// both the matching pattern and filepaths passed to the callback, and is
// typically required on Windows. It usually has no effect on systems where
// forward slash is the path separator, so it is enabled by default.
func TranslateSlashes(enable bool) GlobOption {
	return func(cfg *globConfig) {
		cfg.translateSlashes = enable
	}
}

// WalkIntermediateDirs enables or disables calling the walk function with
// intermediate directory paths (in addition to complete pattern matches).
// Enabling this is needed in order to skip intermediate directories (by
// returning fs.SkipDir) based on custom logic in your walk callback.
// Note that, when enabled, the callback can be called with intermediate
// directories that ultimately do not contain any completely matching paths.
// (e.g. when globbing "fixtures/**/*_test", every subdirectory of "fixtures"
// will be passed to your callback whether or not there is a file named like
// "*_test" within).
// Disabled by default.
func WalkIntermediateDirs(enable bool) GlobOption {
	return func(cfg *globConfig) {
		cfg.walkIntermediateDirs = enable
	}
}

// GoroutineLimit sets a concurrency limit for MultiGlob. By default there is no
// limit. MultiGlob will create at most n worker goroutines unless n <= 0.
func GoroutineLimit(n int) GlobOption {
	return func(cfg *globConfig) {
		cfg.goroutines = n
	}
}
