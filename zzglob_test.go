package zzglob

import (
	"io/fs"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type walkFuncArgs struct {
	Path string
	Err  error
}

func TestGlob(t *testing.T) {
	var got []walkFuncArgs
	err := Glob("fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m", func(path string, d fs.DirEntry, err error) error {
		got = append(got, walkFuncArgs{path, err})
		return nil
	}, true)
	if err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := []walkFuncArgs{
		{Path: "fixtures/a/b/cd/elf/g/j/absurdity/m"},
		{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}
