package zzglob

import (
	"errors"
	"fmt"
)

func parse(pattern string) (string, expression, error) {
	tks := tokenise(pattern)
	seq, _, err := parseSequence(&tks, false)
	// TODO: handle finding the root
	return "", seq, err
}

func parseSequence(tks *tokens, insideAlt bool) (seq sequenceExp, done bool, err error) {
	for {
		t := tks.next()
		if t == nil {
			return seq, true, nil
		}

		switch t := t.(type) {
		case literal:
			seq = append(seq, literalExp(t))

		case punctuation:
			switch t {
			case '/':
				seq = append(seq, pathSepExp{})

			case '*':
				seq = append(seq, starExp{})

			case '⁑':
				seq = append(seq, doubleStarExp{})

			case '?':
				seq = append(seq, questionExp{})

			case '{':
				a, err := parseAlternation(tks)
				if err != nil {
					return nil, false, err
				}
				seq = append(seq, a)

			case '}':
				if insideAlt {
					return seq, true, nil
				}
				seq = append(seq, literalExp('}'))

			case ',':
				if insideAlt {
					return seq, false, nil
				}
				seq = append(seq, literalExp(','))

			case '[':
				c, err := parseCharClass(tks)
				if err != nil {
					return nil, false, err
				}
				seq = append(seq, c)

			default:
				return nil, false, fmt.Errorf("invalid punctuation %c", t)
			}

		default:
			return nil, false, fmt.Errorf("invalid token type %T", t)
		}
	}
}

func parseAlternation(tks *tokens) (alternationExp, error) {
	// TODO: need to handle missing closing brace
	var a alternationExp
	for {
		seq, done, err := parseSequence(tks, true)
		if err != nil {
			return nil, err
		}
		a = append(a, seq)
		if done {
			return a, nil
		}
	}
}

func parseCharClass(tks *tokens) (charClassExp, error) {
	cc := make(charClassExp)
	for {
		t := tks.next()
		if t == nil {
			return nil, errors.New("unterminated char class")
		}
		switch t := t.(type) {
		case literal:
			cc[rune(t)] = struct{}{}

		case punctuation:
			switch t {
			case ']':
				return cc, nil

			default:
				return nil, fmt.Errorf("invalid %c within char class", t)
			}
		}
	}
}
