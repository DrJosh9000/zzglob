// Package zzglob implements a file path walker.
package zzglob

import (
	"errors"
	"io/fs"
)

// Glob globs for files matching the pattern.
func Glob(pattern string, f fs.WalkDirFunc) error {
	return errors.New("unimplemented")
}