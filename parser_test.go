package zzglob

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	input := "abc/{d,e}"
	_, got, err := parse(input)
	if err != nil {
		t.Fatalf("parse(%q) error = %v", input, err)
	}

	want := &state{Out: []edge{{
		Expr: literalExp('a'),
		State: &state{Out: []edge{{
			Expr: literalExp('b'),
			State: &state{Out: []edge{{
				Expr: literalExp('c'),
				State: &state{Out: []edge{{
					Expr: literalExp('/'),
					State: &state{Out: []edge{
						{Expr: literalExp('d'), State: &state{}},
						{Expr: literalExp('e'), State: &state{}},
					}},
				}}},
			}}},
		}}},
	}}}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("parsed diff (-got +want):\n%s", diff)
	}
}
