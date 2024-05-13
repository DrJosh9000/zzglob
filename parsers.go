package zzglob

import (
	"errors"
	"fmt"
	"slices"
)

type parserContext int

const (
	parserInsideNothing parserContext = iota
	parserInsideAlternation
	parserInsideShellExtGlob
	parserInsideBangParen
)

// parseSequence parses a sequence.
func parseSequence(tkns *tokens, pctx parserContext) (start, end *state, endedWith token, err error) {
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
		t, ok := tkns.next()
		if !ok {
			return start, end, 0, nil
		}

		if t >= 0 {
			appendExp(literalExp(t))
			continue
		}

		switch t {
		case tokenStar:
			end.Out = append(end.Out, edge{
				Expr:  starExp{},
				State: end,
			})

		case tokenDoubleStar:
			end.Out = append(end.Out, edge{
				Expr:  doubleStarExp{},
				State: end,
			})

		case tokenQuestion:
			appendExp(questionExp{})

		case tokenOpenBrace:
			ed, err := parseAlternation(tkns, end)
			if err != nil {
				return nil, nil, 0, err
			}
			end = ed

		case tokenCloseBrace:
			if pctx == parserInsideAlternation {
				return start, end, t, nil
			}
			appendExp(literalExp('}'))

		case tokenComma:
			if pctx == parserInsideAlternation {
				return start, end, t, nil
			}
			appendExp(literalExp(','))

		case tokenPipe:
			if pctx == parserInsideShellExtGlob {
				return start, end, t, nil
			}
			appendExp(literalExp('|'))

		case tokenCloseParen:
			if pctx == parserInsideShellExtGlob {
				return start, end, t, nil
			}
			appendExp(literalExp(')'))

		case tokenOpenBracket:
			ed, err := parseCharClass(tkns, end)
			if err != nil {
				return nil, nil, 0, err
			}
			end = ed

		case tokenCloseBracket:
			appendExp(literalExp(']'))

		case tokenBracketCaret:
			ed, err := parseNegatedCharClass(tkns, end)
			if err != nil {
				return nil, nil, 0, err
			}
			end = ed

		case tokenAtParen, tokenStarParen, tokenPlusParen, tokenQuestionParen:
			ed, err := parseShellExtGlob(tkns, end, t)
			if err != nil {
				return nil, nil, 0, err
			}
			end = ed

		case tokenBangParen:
			if pctx == parserInsideBangParen {
				return nil, nil, 0, fmt.Errorf("nested negative extglobs are not supported")
			}
			ed, err := parseBangParenExtGlob(tkns, end)
			if err != nil {
				return nil, nil, 0, err
			}
			end = ed

		default:
			return nil, nil, 0, fmt.Errorf("invalid punctuation %c", t)
		}
	}
}

// parseAlternation appends a branch to the automaton, a sequence in each
// branch, then a merge.
func parseAlternation(tks *tokens, from *state) (end *state, err error) {
	end = &state{}
	for {
		st, ed, done, err := parseSequence(tks, parserInsideAlternation)
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
		case tokenComma:
			continue

		case tokenCloseBrace:
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
		t, ok := tks.next()
		if !ok {
			return nil, errors.New("unterminated char class - missing closing square bracket")
		}
		if t >= 0 {
			from.Out = append(from.Out, edge{
				Expr:  literalExp(t),
				State: end,
			})
			continue
		}

		if t != tokenCloseBracket {
			return nil, fmt.Errorf("invalid %s (%d) within char class", t, t)
		}
		return end, nil
	}
}

// parseNegatedCharClass parses a negated char class. tks should start with the
// the first token following `[^`.
func parseNegatedCharClass(tks *tokens, from *state) (*state, error) {
	runes := make(map[rune]struct{})
	for {
		t, ok := tks.next()
		if !ok {
			return nil, errors.New("unterminated negated char class - missing closing square bracket")
		}
		if t >= 0 {
			runes[rune(t)] = struct{}{}
			continue
		}

		if t != tokenCloseBracket {
			return nil, fmt.Errorf("invalid %s (%d) within negated char class", t, t)
		}
		break
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

// parseShellExtGlob parses any of the shell extglob patterns except !(...).
func parseShellExtGlob(tks *tokens, from *state, kind token) (end *state, err error) {
	end = &state{}

	if kind == tokenQuestionParen || kind == tokenStarParen {
		// ?( = Zero or one; *( = zero or more. => can skip. Add a skip edge.
		from.Out = append(from.Out, edge{Expr: nil, State: end})
	}
	if kind == tokenPlusParen || kind == tokenStarParen {
		// +( = one or more; *( or more. Add a loop edge.
		end.Out = append(end.Out, edge{Expr: nil, State: from})
	}
	// @( just works like an alternation.

	for {
		st, ed, done, err := parseSequence(tks, parserInsideShellExtGlob)
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
		case tokenPipe:
			continue

		case tokenCloseParen:
			return end, nil

		default:
			return nil, errors.New("unterminated extglob - missing closing parenthesis")
		}
	}
}

// parseBangParenExtGlob parses !(...)
func parseBangParenExtGlob(tks *tokens, from *state) (end *state, err error) {
	end = &state{}
	for {
		st, _, done, err := parseSequence(tks, parserInsideBangParen)
		if err != nil {
			return nil, err
		}
		from.Out = append(from.Out, edge{
			Expr:  nil,
			State: st,
		})

		switch done {
		case tokenPipe:
			continue

		case tokenCloseParen:
			return end, nil

		default:
			return nil, errors.New("unterminated extglob - missing closing parenthesis")
		}
	}
}
