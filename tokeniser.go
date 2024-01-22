package zzglob

import (
	"fmt"
	"strings"
)

// Lexer tokens
type (
	literal     rune // not any of the below
	punctuation rune // *, ** (as ⁑), ?, {, }, [, ], ~, or comma
)

func (literal) tokenTag()     {}
func (punctuation) tokenTag() {}

func (l literal) String() string     { return fmt.Sprintf("literal(%q)", rune(l)) }
func (p punctuation) String() string { return fmt.Sprintf("punctuation(%q)", rune(p)) }

type token interface {
	tokenTag()
	String() string
}

type tokens []token

func tokenise(p string, cfg *parseConfig) *tokens {
	// Most tokens are single runes, so preallocate len(p).
	tks := make(tokens, 0, len(p))

	pathSep := '/'
	escapeChar := '\\'
	if cfg.swapSlashes {
		pathSep, escapeChar = escapeChar, pathSep
	}

	// Tokenisation state.
	escape := false        // the previous char was escapeChar
	star := false          // the previous char was *
	insideCC := false      // within a char class
	insideCCFirst := false // first token within a char class

	// Walk through string, producing tokens.
	for _, c := range p {
		// Escaping something?
		if escape {
			// The escapeChar escaped c, so c is a literal.
			escape = false
			tks = append(tks, literal(c))
			continue
		}

		// Within a char class? No escaping required, other than ].
		// But the user can optionally escape everything, and must if they want
		// ], so escape is higher priority.
		if insideCC {
			switch c {
			case escapeChar:
				if cfg.allowEscaping { // Start of escape
					escape = true
				} else {
					tks = append(tks, literal(escapeChar))
				}

			case ']':
				// End of cc
				tks = append(tks, punctuation(']'))
				insideCC = false

			case '^':
				// Negated char class only if ^ is first token inside [ ]
				if insideCCFirst {
					tks = append(tks, punctuation('^'))
				} else {
					tks = append(tks, literal('^'))
				}

			default:
				tks = append(tks, literal(c))
			}

			insideCCFirst = false
			continue
		}

		// Wishing upon a *?
		if star {
			star = false
			if c == '*' {
				tks = append(tks, punctuation('⁑'))
				continue
			}

			// The previous char was a *, but this one isn't.
			// Emit *, then process c normally.
			tks = append(tks, punctuation('*'))
		}

		switch c {
		case '*': // previous char is not *
			// It could be a * or **.
			if cfg.allowStar {
				if cfg.allowDoubleStar {
					star = true
				} else {
					tks = append(tks, punctuation('*'))
				}
			} else {
				tks = append(tks, literal('*'))
			}

		case escapeChar:
			if !cfg.allowEscaping {
				tks = append(tks, literal(escapeChar))
				break
			}
			// Next char is escaped.
			escape = true

		case pathSep:
			// Always represent the path separator with / for consistency
			// with io/fs.
			tks = append(tks, literal('/'))

		case '~':
			if cfg.expandTilde {
				tks = append(tks, punctuation('~'))
			} else {
				tks = append(tks, literal('~'))
			}

		case '[':
			if cfg.allowCharClass {
				insideCC = true
				insideCCFirst = true
				tks = append(tks, punctuation('['))
			} else {
				tks = append(tks, literal('['))
			}

		case '?':
			if cfg.allowQuestion {
				tks = append(tks, punctuation('?'))
			} else {
				tks = append(tks, literal('?'))
			}

		case ']':
			// We only get here if insideCC is false...
			tks = append(tks, literal(']'))

		case '{', '}', ',':
			if cfg.allowAlternation {
				tks = append(tks, punctuation(c))
			} else {
				tks = append(tks, literal(c))
			}

		default:
			// It's a literal.
			tks = append(tks, literal(c))
		}
	}

	// Escape or * at end of string?
	if escape {
		tks = append(tks, literal(escapeChar))
	}
	if star {
		tks = append(tks, punctuation('*'))
	}

	return &tks
}

// next uses a pointer to a slice as a consuming reader.
func (r *tokens) next() any {
	if r == nil || len(*r) == 0 {
		return nil
	}
	defer func() { *r = (*r)[1:] }()
	return (*r)[0]
}

// allLiteral returns a string consisting of all tokens runic equivalents.
// r must consist solely of literals, otherwise it returns the empty string.
func (r tokens) allLiteral() string {
	b := strings.Builder{}
	b.Grow(len(r))
	for _, t := range r {
		t, ok := t.(literal)
		if !ok {
			return ""
		}
		b.WriteRune(rune(t))
	}
	return b.String()
}
