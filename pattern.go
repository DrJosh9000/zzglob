package zzglob

import (
	"fmt"
	"io"
)

// Pattern is a parsed glob pattern.
type Pattern struct {
	root    string
	initial *state
}

// Parse parses a pattern.
func Parse(pattern string, opts ...ParseOption) (*Pattern, error) {
	cfg := defaultParseConfig
	for _, o := range opts {
		o(&cfg)
	}

	// tokenise classifies each rune as literal or punctuation, interprets
	// escape chars, etc.
	tks := tokenise(pattern, &cfg)

	// Preprocessing, for example replace ~/ with homedir.
	*tks = preprocess(*tks)

	// If the pattern is all literals, then it's a specific path.
	if root := tks.allLiteral(); root != "" {
		return &Pattern{
			root:    root,
			initial: nil,
		}, nil
	}

	// Find the root of the path. This is where directory walking starts.
	root := findRoot(tks)

	// Convert the rest of the sequence into a state machine.
	initial, terminal, _, err := parseSequence(tks, parserInsideNothing)
	if err != nil {
		return nil, err
	}

	// The terminal state is accepting.
	terminal.Accept = true

	// Remove redundant nil edges, where possible. This should only ever remove
	// edges and possibly redundant intermediate states.
	reduce(initial)

	// Done! Here's the machine.
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

// WriteDot writes a digraph representing the state machine to the writer
// (in GraphViz syntax).
func (p *Pattern) WriteDot(w io.Writer, hilite stateSet) error {
	if _, err := fmt.Fprintln(w, "digraph {\n\trankdir=LR;"); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "\tinitial [label=\"\", style=invis];"); err != nil {
		return err
	}

	if p.initial == nil {
		if _, err := fmt.Fprintln(w, "\tterminal [label=\"\", shape=doublecircle];"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "\tinitial -> terminal [label=\"%s\"];\n", p.root); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "}"); err != nil {
			return err
		}
		return nil
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
		if s.Accept {
			shape = "doublecircle"
		}
		fill := "white"
		if _, ok := hilite[s]; ok {
			fill = "green"
		}
		if _, err := fmt.Fprintf(w, "\tstate_%p [label=\"\", shape=%s, style=filled, fillcolor=%s];\n", s, shape, fill); err != nil {
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

// reduce tries to safely eliminate any edges with nil expression that it can
// find. "Safely" means both correctness (not changing which inputs the
// machine accepts or rejects) and complexity (e.g. not adding O(n^2) edges to
// replace the nil edges it eliminates).
func reduce(initial *state) {
	seen := make(map[*state]bool)
	q := []*state{initial}
	for len(q) > 0 {
		s := q[0]
		q = q[1:]

		if seen[s] {
			continue
		}
		seen[s] = true

		for i := range s.Out {
			e := &s.Out[i]

			// These optimisations only apply if the destination state is valid
			// and has out-degree 1.
			for {
				// If e has nil Expr, then replace both the expression and
				// target of e with the next edge:
				//
				// s --e(<nil>)--> s' --e'--> s''
				//   becomes
				// s --e'--> s''
				if e.State != nil && len(e.State.Out) == 1 && e.Expr == nil {
					*e = e.State.Out[0]
					continue
				}

				// If the next edge has nil expression, then replace the target
				// state of e with the target of that subsequent edge.
				//
				// s --e--> s' --e'(<nil>)--> s''
				//   becomes
				// s --e--> s''
				if e.State != nil && len(e.State.Out) == 1 && e.State.Out[0].Expr == nil {
					e.State = e.State.Out[0].State
					continue
				}

				break
			}
		}

		for _, e := range s.Out {
			if !seen[e.State] {
				q = append(q, e.State)
			}
		}
	}
}
