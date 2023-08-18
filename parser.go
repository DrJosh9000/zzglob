package zzglob

import (
	"errors"
	"fmt"
)

// parse converts a pattern into a finite automaton.
func parse(pattern string) (string, *state, error) {
	tks := tokenise(pattern)
	// var root strings.Builder
	// for _, t := range tks {
	// 	l, ok := t.(literal)
	// 	if !ok {
	// 		break
	// 	}
	// 	root.WriteRune(rune(l))
	// }
	// if root.String() == pattern {
	// 	return pattern, nil, nil
	// }

	// TODO: handle finding the root
	n, _, _, err := parseSequence(&tks, false)
	if err != nil {
		return "", nil, err
	}
	reduce(n)
	return "", n, nil
}

// reduce recursively eliminates any edges with nil expression.
func reduce(n *state) {
	var enew []edge
	for _, e := range n.Out {
		if e.State == nil {
			continue
		}
		reduce(e.State)
		if e.Expr == nil {
			// e is an expressionless edge. Using the next node's edges.
			enew = append(enew, e.State.Out...)
			continue
		}
		enew = append(enew, e)
	}
	n.Out = enew
}

// parseSequence parses a sequence into a finite automaton.
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
				appendExp(starExp{})

			case '⁑':
				appendExp(doubleStarExp{})

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
		from.Out = append(from.Out, st.Out...)
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

// parseCharClass is like parseAlternation.
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
