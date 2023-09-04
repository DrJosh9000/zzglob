package zzglob

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	input := "x/y/z/*abc/{d,e}"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) error = %v", input, err)
	}

	loop := &state{}

	want := &Pattern{
		root: "x/y/z",
		initial: &state{
			Out: []edge{{
				Expr:  literalExp('/'),
				State: loop,
			}},
		},
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
		"a/b*c/d?e/{f,g}/[ij]/**/k",
	}
	for _, pattern := range tests {
		p, err := Parse(pattern)
		if err != nil {
			t.Fatalf("Parse(%q) error = %v", pattern, err)
		}
		if err := p.WriteDot(io.Discard, nil); err != nil {
			t.Errorf("(%q).WriteDot(io.Discard) = %v", pattern, err)
		}
	}
}
