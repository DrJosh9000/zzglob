package zzglob

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func mustMultiParse(t *testing.T, patterns ...string) []*Pattern {
	t.Helper()
	patts := make([]*Pattern, 0, len(patterns))
	for _, patt := range patterns {
		p, err := Parse(patt)
		if err != nil {
			t.Fatalf("Parse(%q) = %v", patt, err)
		}
		patts = append(patts, p)
	}
	return patts
}

func TestMultiGlob_SinglePattern(t *testing.T) {
	patterns := mustMultiParse(t,
		"fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m",
	)

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b/cd/elf/g/j/absurdity/m"},
			{Path: "fixtures/a/b/cid/erf/h/k/m"},
			{Path: "fixtures/a/b/cid/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_MultiplePatterns_DifferentRoots(t *testing.T) {
	patterns := mustMultiParse(t,
		"fixtures/a/b/cid/**/m",
		"fixtures/a/b/cod/**/m",
	)

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b/cid/erf/h/k/m"},
			{Path: "fixtures/a/b/cid/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cid/erf/i/m"},
			{Path: "fixtures/a/b/cid/erf/i/n/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cod/erf/i/m"},
			{Path: "fixtures/a/b/cod/erf/i/n/m"},
		},
	}

	got.sortCalls()

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_MultiplePatterns_DifferentRoots_WalkIntermediateDirs(t *testing.T) {
	patterns := mustMultiParse(t,
		"fixtures/a/b/cid/**/m",
		"fixtures/a/b/cod/**/m",
	)

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, traceLogOpt, WalkIntermediateDirs(true)); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b/cid"},
			{Path: "fixtures/a/b/cid/erf"},
			{Path: "fixtures/a/b/cid/erf/h"},
			{Path: "fixtures/a/b/cid/erf/h/k"},
			{Path: "fixtures/a/b/cid/erf/h/k/m"},
			{Path: "fixtures/a/b/cid/erf/h/k/n"},
			{Path: "fixtures/a/b/cid/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cid/erf/i"},
			{Path: "fixtures/a/b/cid/erf/i/m"},
			{Path: "fixtures/a/b/cid/erf/i/n"},
			{Path: "fixtures/a/b/cid/erf/i/n/m"},
			{Path: "fixtures/a/b/cod"},
			{Path: "fixtures/a/b/cod/erf"},
			{Path: "fixtures/a/b/cod/erf/h"},
			{Path: "fixtures/a/b/cod/erf/h/k"},
			{Path: "fixtures/a/b/cod/erf/h/k/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/n"},
			{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cod/erf/i"},
			{Path: "fixtures/a/b/cod/erf/i/m"},
			{Path: "fixtures/a/b/cod/erf/i/n"},
			{Path: "fixtures/a/b/cod/erf/i/n/m"},
		},
	}

	got.sortCalls()

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_MultiplePatterns_SameRoot(t *testing.T) {
	patterns := mustMultiParse(t,
		"fixtures/a/b/c{i}d/**/m",
		"fixtures/a/b/c{o}d/**/m",
	)

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b/cid/erf/h/k/m"},
			{Path: "fixtures/a/b/cid/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cid/erf/i/m"},
			{Path: "fixtures/a/b/cid/erf/i/n/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cod/erf/i/m"},
			{Path: "fixtures/a/b/cod/erf/i/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_TraverseSymlinksDisabled(t *testing.T) {
	patterns := mustMultiParse(t,
		"fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m",
	)

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, TraverseSymlinks(false), traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b/cd/elf/g/j/absurdity/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_Absolute(t *testing.T) {
	src := "fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	pattern, err := filepath.Abs(filepath.FromSlash(src))
	if err != nil {
		t.Fatalf("filepath.Abs(%q) error = %v", src, err)
	}
	pattern = filepath.ToSlash(pattern)
	base := strings.TrimSuffix(pattern, src)

	patterns := mustMultiParse(t, pattern)

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: base + "fixtures/a/b/cd/elf/g/j/absurdity/m"},
			{Path: base + "fixtures/a/b/cid/erf/h/k/m"},
			{Path: base + "fixtures/a/b/cid/erf/h/k/n/m"},
			{Path: base + "fixtures/a/b/cod/erf/h/k/m"},
			{Path: base + "fixtures/a/b/cod/erf/h/k/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_SpecificPath(t *testing.T) {
	patterns := mustMultiParse(t,
		"fixtures/a/b/cod/erf/h/k/n/m",
	)

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, TraverseSymlinks(false), traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_EmptyRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	defer os.Chdir(wd)
	os.Chdir("fixtures")

	patterns := mustMultiParse(t, "**/m")

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "a/b/cad/m"},
			{Path: "a/b/cd/elf/g/j/absurdity/m"},
			{Path: "a/b/cid/erf/h/k/m"},
			{Path: "a/b/cid/erf/h/k/n/m"},
			{Path: "a/b/cid/erf/i/m"},
			{Path: "a/b/cid/erf/i/n/m"},
			{Path: "a/b/cod/erf/h/k/m"},
			{Path: "a/b/cod/erf/h/k/n/m"},
			{Path: "a/b/cod/erf/i/m"},
			{Path: "a/b/cod/erf/i/n/m"},
			{Path: "m"},
			{
				Path: "spec/borked",
				Err:  &fs.PathError{Op: "stat", Path: "spec/borked", Err: syscall.ENOENT},
			},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_WithFilesystem(t *testing.T) {
	patterns := mustMultiParse(t, "a/b/c*d/e?f/[ghi]/{j,k,l}/**/m")

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, WithFilesystem(os.DirFS("fixtures")), traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "a/b/cd/elf/g/j/absurdity/m"},
			{Path: "a/b/cid/erf/h/k/m"},
			{Path: "a/b/cid/erf/h/k/n/m"},
			{Path: "a/b/cod/erf/h/k/m"},
			{Path: "a/b/cod/erf/h/k/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestMultiGlob_SpecificPath_WithFilesystem(t *testing.T) {
	patterns := mustMultiParse(t, "a/b/cod/erf/h/k/n/m")

	var got walkFuncCalls
	if err := MultiGlob(context.Background(), patterns, got.walkFunc, WithFilesystem(os.DirFS("fixtures")), traceLogOpt); err != nil {
		t.Fatalf("MultiGlob(...) = %v", err)
	}
	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "a/b/cod/erf/h/k/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}
