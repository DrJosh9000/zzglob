package zzglob

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	input := "x/y/z/*abc/{d,e}"
	gotRoot, gotState, err := parse(input)
	if err != nil {
		t.Fatalf("parse(%q) error = %v", input, err)
	}

	wantRoot := "x/y/z/"

	// stars form loops
	wantState := new(state)
	wantState.Out = append(wantState.Out, edge{
		Expr:  starExp{},
		State: wantState,
	})
	wantState.Out = append(wantState.Out, edge{
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
	})

	if diff := cmp.Diff(gotRoot, wantRoot); diff != "" {
		t.Errorf("parsed state diff (-got +want):\n%s", diff)
	}

	if diff := cmp.Diff(gotState, wantState); diff != "" {
		t.Errorf("parsed state diff (-got +want):\n%s", diff)
	}
}
