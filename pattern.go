package zzglob

import (
	"fmt"
	"io"
)

// Pattern is a glob pattern.
type Pattern struct {
	root    string
	initial *state
}

// Parse converts a pattern into a finite automaton.
func Parse(pattern string) (*Pattern, error) {
	// tokenise classifies each rune as literal or punctuation
	tks := tokenise(pattern)

	// Find the root of the path. This is where directory walking starts.
	root := findRoot(tks)

	// Convert the rest of the sequence into a DFA.
	initial, _, _, err := parseSequence(tks, false)
	if err != nil {
		return nil, err
	}
	reduce(initial)
	return &Pattern{
		root:    root,
		initial: initial,
	}, nil
}

// MustParse calls Parse, and panics if unable to parse the pattern.
func MustParse(pattern string) *Pattern {
	p, err := Parse(pattern)
	if err != nil {
		panic(err)
	}
	return p
}

// WriteDot writes a digraph representing the automaton to the writer
// (in GraphViz syntax).
func (p *Pattern) WriteDot(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "digraph {\n\trankdir=LR;"); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "\tinitial [label=\"\", style=\"invis\"];"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\tinitial -> state_%p [label=\"%s\"];\n", p.initial, p.root); err != nil {
		return err
	}

	seen := make(map[*state]bool)
	q := []*state{p.initial}
	for len(q) > 0 {
		s := q[0]
		q = q[1:]

		if seen[s] {
			continue
		}
		seen[s] = true

		shape := "circle"
		if s.terminal() {
			shape = "doublecircle"
		}
		if _, err := fmt.Fprintf(w, "\tstate_%p [label=\"\", shape=\"%s\"];\n", s, shape); err != nil {
			return err
		}
		for _, e := range s.Out {
			if _, err := fmt.Fprintf(w, "\tstate_%p -> state_%p [label=\"%v\"];\n", s, e.State, e.Expr); err != nil {
				return err
			}
			if seen[e.State] {
				continue
			}
			q = append(q, e.State)
		}
	}
	if _, err := fmt.Fprintln(w, "}"); err != nil {
		return err
	}
	return nil
}
