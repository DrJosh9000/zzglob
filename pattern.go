package zzglob

import (
	"errors"
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
	initial, terminal, _, err := parseSequence(tks, false)
	if err != nil {
		return nil, err
	}

	// The terminal state is terminal.
	terminal.Terminal = true

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
		if s.Terminal {
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

// parseSequence parses a sequence.
func parseSequence(tkns *tokens, insideAlt bool) (start, end *state, endedWith token, err error) {
	start = &state{}
	end = start
	appendExp := func(e expression) {
		next := &state{}
		end.Out = append(end.Out, edge{
			Expr:  e,
			State: next,
		})
		end = next
	}

	for {
		t := tkns.next()
		if t == nil {
			return start, end, nil, nil
		}

		switch t := t.(type) {
		case literal:
			appendExp(literalExp(t))

		case punctuation:
			switch t {
			case '*':
				end.Out = append(end.Out, edge{
					Expr:  starExp{},
					State: end,
				})

			case '⁑':
				end.Out = append(end.Out, edge{
					Expr:  doubleStarExp{},
					State: end,
				})

			case '?':
				appendExp(questionExp{})

			case '{':
				ed, err := parseAlternation(tkns, end)
				if err != nil {
					return nil, nil, nil, err
				}
				end = ed

			case '}':
				if insideAlt {
					return start, end, t, nil
				}
				appendExp(literalExp('}'))

			case ',':
				if insideAlt {
					return start, end, t, nil
				}
				appendExp(literalExp(','))

			case '[':
				ed, err := parseCharClass(tkns, end)
				if err != nil {
					return nil, nil, nil, err
				}
				end = ed

			default:
				return nil, nil, nil, fmt.Errorf("invalid punctuation %c", t)
			}

		default:
			return nil, nil, nil, fmt.Errorf("invalid token type %T", t)
		}
	}
}

// parseAlternation appends a branch to the automaton, a sequence in each
// branch, then a merge.
func parseAlternation(tks *tokens, from *state) (end *state, err error) {
	end = &state{}
	for {
		st, ed, done, err := parseSequence(tks, true)
		if err != nil {
			return nil, err
		}
		from.Out = append(from.Out, edge{
			Expr:  nil,
			State: st,
		})
		ed.Out = append(ed.Out, edge{
			Expr:  nil,
			State: end,
		})

		switch done {
		case punctuation(','):
			continue

		case punctuation('}'):
			return end, nil

		default:
			return nil, errors.New("unterminated alternation - missing closing brace")
		}
	}
}

// parseCharClass is like parseAlternation, except each branch only matches
// exactly one character.
func parseCharClass(tks *tokens, from *state) (end *state, err error) {
	end = &state{}
	for {
		t := tks.next()
		if t == nil {
			return nil, errors.New("unterminated char class - missing closing square bracket")
		}
		switch t := t.(type) {
		case literal:
			from.Out = append(from.Out, edge{
				Expr:  literalExp(t),
				State: end,
			})

		case punctuation:
			switch t {
			case ']':
				return end, nil

			default:
				return nil, fmt.Errorf("invalid %c within char class", t)
			}
		}
	}
}
