package zzglob

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var traceLogOpt GlobOption = nil // WithTraceLogs(os.Stderr)

type walkFuncArgs struct {
	Path string
	Err  error
}

type walkFuncCalls struct {
	mu    sync.Mutex
	calls []walkFuncArgs
}

func (c *walkFuncCalls) walkFunc(path string, d fs.DirEntry, err error) error {
	c.mu.Lock()
	c.calls = append(c.calls, walkFuncArgs{path, err})
	c.mu.Unlock()
	return nil
}

func (c *walkFuncCalls) sortCalls() {
	c.mu.Lock()
	defer c.mu.Unlock()
	sort.Slice(c.calls, func(i, j int) bool {
		// Only sort path for now
		return c.calls[i].Path < c.calls[j].Path
	})
}

func TestGlob(t *testing.T) {
	pattern := "fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("Glob(...) = %v", err)
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

func TestGlob_RubySpecs(t *testing.T) {
	pattern := "fixtures/**/*_spec.rb"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/spec/bar_spec.rb"},
			{Path: "fixtures/spec/foo_spec.rb"},
			{Path: "fixtures/spec/model/qux_spec.rb"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestGlob_WalkIntermediateDirs(t *testing.T) {
	pattern := "fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt, WalkIntermediateDirs(true)); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b"},
			{Path: "fixtures/a/b/cad"},
			{Path: "fixtures/a/b/cd"},
			{Path: "fixtures/a/b/cd/elf"},
			{Path: "fixtures/a/b/cd/elf/g"},
			{Path: "fixtures/a/b/cd/elf/g/j"},
			{Path: "fixtures/a/b/cd/elf/g/j/absurdity"},
			{Path: "fixtures/a/b/cd/elf/g/j/absurdity/m"},
			{Path: "fixtures/a/b/cid"},
			{Path: "fixtures/a/b/cid/erf"},
			{Path: "fixtures/a/b/cid/erf/h"},
			{Path: "fixtures/a/b/cid/erf/h/k"},
			{Path: "fixtures/a/b/cid/erf/h/k/m"},
			{Path: "fixtures/a/b/cid/erf/h/k/n"},
			{Path: "fixtures/a/b/cid/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cid/erf/i"},
			{Path: "fixtures/a/b/cod"},
			{Path: "fixtures/a/b/cod/erf"},
			{Path: "fixtures/a/b/cod/erf/h"},
			{Path: "fixtures/a/b/cod/erf/h/k"},
			{Path: "fixtures/a/b/cod/erf/h/k/m"},
			{Path: "fixtures/a/b/cod/erf/h/k/n"},
			{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cod/erf/i"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestGlob_GoTests_WalkIntermediateDirs(t *testing.T) {
	pattern := "fixtures/**/*_test.go"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt, WalkIntermediateDirs(true)); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures"},
			{Path: "fixtures/a"},
			{Path: "fixtures/a/b"},
			{Path: "fixtures/a/b/cad"},
			{Path: "fixtures/a/b/cd"},
			{Path: "fixtures/a/b/cd/elf"},
			{Path: "fixtures/a/b/cd/elf/g"},
			{Path: "fixtures/a/b/cd/elf/g/j"},
			{Path: "fixtures/a/b/cd/elf/g/j/absurdity"},
			{Path: "fixtures/a/b/cid"},
			{Path: "fixtures/a/b/cid/erf"},
			{Path: "fixtures/a/b/cid/erf/h"},
			{Path: "fixtures/a/b/cid/erf/h/k"},
			{Path: "fixtures/a/b/cid/erf/h/k/n"},
			{Path: "fixtures/a/b/cid/erf/i"},
			{Path: "fixtures/a/b/cid/erf/i/n"},
			{Path: "fixtures/a/b/cod"},
			{Path: "fixtures/a/b/cod/erf"},
			{Path: "fixtures/a/b/cod/erf/h"},
			{Path: "fixtures/a/b/cod/erf/h/k"},
			{Path: "fixtures/a/b/cod/erf/h/k/n"},
			{Path: "fixtures/a/b/cod/erf/i"},
			{Path: "fixtures/a/b/cod/erf/i/n"},
			{Path: "fixtures/spec"},
			{Path: "fixtures/spec/cmd"},
			{Path: "fixtures/spec/cmd/cmd_test.go"},
			{Path: "fixtures/spec/foo_test.go"},
			{Path: "fixtures/spec/model"},
			{Path: "fixtures/spec/snake"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestGlob_SymlinkInSymlink(t *testing.T) {
	// cid       -> symlink to cod
	// cod/erf/i -> symlink to cod/erf/h/k
	// So
	// ci/erf/i -> cod/erf/h/k ... right?
	pattern := "fixtures/a/b/c{i}d/**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		calls: []walkFuncArgs{
			{Path: "fixtures/a/b/cid/erf/h/k/m"},
			{Path: "fixtures/a/b/cid/erf/h/k/n/m"},
			{Path: "fixtures/a/b/cid/erf/i/m"},
			{Path: "fixtures/a/b/cid/erf/i/n/m"},
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestGlob_TraverseSymlinksDisabled(t *testing.T) {
	pattern := "fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, TraverseSymlinks(false), traceLogOpt); err != nil {
		t.Fatalf("Glob(...) = %v", err)
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

func TestGlob_Absolute(t *testing.T) {
	src := "fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	pattern, err := filepath.Abs(filepath.FromSlash(src))
	if err != nil {
		t.Fatalf("filepath.Abs(%q) error = %v", src, err)
	}
	pattern = filepath.ToSlash(pattern)
	base := strings.TrimSuffix(pattern, src)

	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("Glob(...) = %v", err)
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

func TestGlob_SpecificPath(t *testing.T) {
	pattern := "fixtures/a/b/cod/erf/h/k/n/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("Glob(...) = %v", err)
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

func TestGlob_EmptyRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	defer os.Chdir(wd)
	os.Chdir("fixtures")

	pattern := "**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt); err != nil {
		t.Fatalf("Glob(...) = %v", err)
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
		},
	}

	if diff := cmp.Diff(got.calls, want.calls); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestGlob_WithFilesystem(t *testing.T) {
	pattern := "a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt, WithFilesystem(os.DirFS("fixtures"))); err != nil {
		t.Fatalf("Glob(...) = %v", err)
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

func TestGlob_SpecificPath_WithFilesystem(t *testing.T) {
	pattern := "a/b/cod/erf/h/k/n/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, traceLogOpt, WithFilesystem(os.DirFS("fixtures"))); err != nil {
		t.Fatalf("Glob(...) = %v", err)
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
