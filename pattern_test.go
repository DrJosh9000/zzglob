package zzglob

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	// Because this tests the parsed output for a specific input:
	// disable SwapSlashes in case this test is ever run on Windows
	input := "x/y/z/*abc/{d,e}"
	got, err := Parse(input, WithSwapSlashes(false))
	if err != nil {
		t.Fatalf("Parse(%q) error = %v", input, err)
	}

	loop := &state{}

	want := &Pattern{
		root:    "x/y/z/",
		initial: loop,
	}

	// stars form loops
	loop.Out = []edge{
		{
			Expr:  starExp{},
			State: loop,
		},
		{
			Expr: literalExp('a'),
			State: &state{Out: []edge{{
				Expr: literalExp('b'),
				State: &state{Out: []edge{{
					Expr: literalExp('c'),
					State: &state{Out: []edge{{
						Expr: literalExp('/'),
						State: &state{Out: []edge{
							{
								Expr:  literalExp('d'),
								State: &state{Terminal: true},
							},
							{
								Expr:  literalExp('e'),
								State: &state{Terminal: true},
							},
						}},
					}}},
				}}},
			}}},
		},
	}

	if diff := cmp.Diff(got.root, want.root); diff != "" {
		t.Errorf("Pattern root diff (-got +want):\n%s", diff)
	}

	if diff := cmp.Diff(got.initial, want.initial); diff != "" {
		t.Errorf("Pattern initial diff (-got +want):\n%s", diff)
	}
}

func TestWriteDotSmoke(t *testing.T) {
	tests := []string{
		"a/b",
		"a/b*c/d?e/{f,g}/[ij]/**/[^k]l",
	}
	for _, pattern := range tests {
		p, err := Parse(pattern, WithSwapSlashes(false))
		if err != nil {
			t.Fatalf("Parse(%q) error = %v", pattern, err)
		}
		if err := p.WriteDot(io.Discard, nil); err != nil {
			t.Errorf("(%q).WriteDot(io.Discard) = %v", pattern, err)
		}
	}
}

func TestParse_SwapSlashes(t *testing.T) {
	// Contains no operators - slash translation only
	src := `C:\Windows\Media\Passport.mid`
	got, err := Parse(src, WithSwapSlashes(true))
	if err != nil {
		t.Errorf("Parse(%q, WithSwapSlashes(true)) error = %v", src, err)
	}

	want := "C:/Windows/Media/Passport.mid"
	if got.root != want {
		t.Errorf("got.root = %q, want %q", got.root, want)
	}
}

func FuzzParseMatch(f *testing.F) {
	f.Fuzz(func(t *testing.T, pattern, path string,
		allowEscaping, allowQuestion, allowStar, allowDoubleStar, allowAlternation, allowCharClass, swapSlashes, expandTilde bool) {
		p, err := Parse(
			pattern,
			AllowEscaping(allowEscaping),
			AllowQuestion(allowQuestion),
			AllowStar(allowStar),
			AllowDoubleStar(allowDoubleStar),
			AllowAlternation(allowAlternation),
			AllowCharClass(allowCharClass),
			WithSwapSlashes(swapSlashes),
			ExpandTilde(expandTilde),
		)

		if err != nil {
			return
		}
		_ = p.Match(path)
	})
}
