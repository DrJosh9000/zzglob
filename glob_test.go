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
	pattern := "fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got []walkFuncArgs
	if err := p.Glob(func(path string, d fs.DirEntry, err error) error {
		got = append(got, walkFuncArgs{path, err})
		return nil
	}, true); err != nil {
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
