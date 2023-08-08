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

	want := &node{Out: []edge{{
		Expr: literalExp('a'),
		Node: &node{Out: []edge{{
			Expr: literalExp('b'),
			Node: &node{Out: []edge{{
				Expr: literalExp('c'),
				Node: &node{Out: []edge{{
					Expr: pathSepExp{},
					Node: &node{Out: []edge{
						{Expr: literalExp('d'), Node: &node{}},
						{Expr: literalExp('e'), Node: &node{}},
					}},
				}}},
			}}},
		}}},
	}}}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("parsed diff (-got +want):\n%s", diff)
	}
}
