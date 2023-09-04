package zzglob

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type walkFuncArgs struct {
	Path string
	Err  error
}

type walkFuncCalls []walkFuncArgs

func (c *walkFuncCalls) walkFunc(path string, d fs.DirEntry, err error) error {
	*c = append(*c, walkFuncArgs{path, err})
	return nil
}

func TestGlob(t *testing.T) {
	pattern := "fixtures/a/b/c*d/e?f/[ghi]/{j,k,l}/**/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, WithTraceLogs(os.Stderr)); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		{Path: "fixtures/a/b/cd/elf/g/j/absurdity/m"},
		{Path: "fixtures/a/b/cid/erf/h/k/m"},
		{Path: "fixtures/a/b/cid/erf/h/k/n/m"},
		{Path: "fixtures/a/b/cod/erf/h/k/m"},
		{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
	}

	if diff := cmp.Diff(got, want); diff != "" {
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
	if err := p.Glob(got.walkFunc, TraverseSymlinks(false), WithTraceLogs(os.Stderr)); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		{Path: "fixtures/a/b/cd/elf/g/j/absurdity/m"},
		{Path: "fixtures/a/b/cod/erf/h/k/m"},
		{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
	}

	if diff := cmp.Diff(got, want); diff != "" {
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
	if err := p.Glob(got.walkFunc, WithTraceLogs(os.Stderr)); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		{Path: base + "fixtures/a/b/cd/elf/g/j/absurdity/m"},
		{Path: base + "fixtures/a/b/cid/erf/h/k/m"},
		{Path: base + "fixtures/a/b/cid/erf/h/k/n/m"},
		{Path: base + "fixtures/a/b/cod/erf/h/k/m"},
		{Path: base + "fixtures/a/b/cod/erf/h/k/n/m"},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}

func TestGlob_SingleFile(t *testing.T) {
	pattern := "fixtures/a/b/cod/erf/h/k/n/m"
	p, err := Parse(pattern)
	if err != nil {
		t.Fatalf("Parse(%q) = %v", pattern, err)
	}

	var got walkFuncCalls
	if err := p.Glob(got.walkFunc, WithTraceLogs(os.Stderr)); err != nil {
		t.Fatalf("Glob(...) = %v", err)
	}

	want := walkFuncCalls{
		{Path: "fixtures/a/b/cod/erf/h/k/n/m"},
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("walked paths diff (-got +want):\n%s", diff)
	}
}
