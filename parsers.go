package zzglob

import (
	"errors"
	"fmt"
	"slices"
)

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

			case punctDoubleStar:
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

			case ']':
				appendExp(literalExp(']'))

			case punctBracketCaret:
				ed, err := parseNegatedCharClass(tkns, end)
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
			if t != ']' {
				return nil, fmt.Errorf("invalid %c within char class", t)
			}
			return end, nil
		}
	}
}

// parseNegatedCharClass parses a negated char class. tks should start with the
// the first token following `[^`.
func parseNegatedCharClass(tks *tokens, from *state) (*state, error) {
	runes := make(map[rune]struct{})
	for {
		t := tks.next()
		if t == nil {
			return nil, errors.New("unterminated negated char class - missing closing square bracket")
		}
		switch t := t.(type) {
		case literal:
			runes[rune(t)] = struct{}{}

		case punctuation:
			if t != ']' {
				return nil, fmt.Errorf("invalid %c within negated char class", t)
			}
			end := &state{}

			expr := make(negatedCCExp, 0, len(runes))
			for r := range runes {
				expr = append(expr, r)
			}
			if len(expr) == 1 {
				from.Out = append(from.Out, edge{
					Expr:  negatedLiteralExp(expr[0]),
					State: end,
				})
				return end, nil
			}

			slices.Sort(expr)

			from.Out = append(from.Out, edge{
				Expr:  expr,
				State: end,
			})
			return end, nil
		}
	}
}
