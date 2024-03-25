package zzglob

import (
	"fmt"
	"strings"
)

// Lexer tokens
type (
	literal     rune // a token that means the rune that it is
	punctuation rune // a token that probably has a special meaning

	// Examples of punctuation are *, **, ?, {, }, [, ], ~, or comma.
	// These are represented as themselves, except for ** which is represented
	// as ⁑.
	// With shell extglob enabled, this also includes ?(, +(, *(, @(, !(, and ),
	// and the | character used as a separator within those.
)

// Here's how to squeeze "multiple rune" punctuation into a single rune:
// consts with special values.
const (
	punctDoubleStar    punctuation = '⁑'  // **
	punctQuestionParen punctuation = -101 // ?(
	punctPlusParen     punctuation = -102 // +(
	punctStarParen     punctuation = -103 // *(
	punctAtParen       punctuation = -104 // @(
	punctBangParen     punctuation = -105 // !(
)

func (literal) tokenTag()     {}
func (punctuation) tokenTag() {}

func (l literal) String() string { return fmt.Sprintf("literal(%q)", rune(l)) }

func (p punctuation) String() string {
	switch p {
	case punctDoubleStar:
		return `punctuation("**")`
	case punctQuestionParen:
		return `punctuation("?(")`
	case punctPlusParen:
		return `punctuation("+(")`
	case punctStarParen:
		return `punctuation("*(")`
	case punctAtParen:
		return `punctuation("@(")`
	case punctBangParen:
		return `punctuation("!(")`
	}
	return fmt.Sprintf("punctuation(%q)", rune(p))
}

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
	// prev is the previous rune read, but only where that rune influences the
	// interpretation of the next rune, e.g. \ or *. Otherwise it is 0.
	var prev rune
	insideCC := false      // within a char class
	insideCCFirst := false // first token within a char class

	// Walk through string, producing tokens.
	for _, c := range p {
		// Escaping something?
		if prev == escapeChar {
			// The escapeChar escaped c, so c is a literal.
			prev = 0
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
					prev = escapeChar
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

		// Possible double-rune token?
		switch prev {
		case '*':
			// Could be *, **, or *( depending on options
			prev = 0
			switch {
			case c == '*' && cfg.allowDoubleStar:
				tks = append(tks, punctDoubleStar)
				continue

			case c == '(' && cfg.enableShellExtGlob:
				tks = append(tks, punctStarParen)
				continue

			default:
				// The previous char was * with possible extra meaning, but c
				// doesn't make it anything special.
				// Emit *, then process c normally.
				if cfg.allowStar {
					tks = append(tks, punctuation('*'))
				} else {
					tks = append(tks, literal('*'))
				}
			}

		case '?':
			// Could be ? or ?( if shell extglob is enabled
			prev = 0
			if c == '(' && cfg.enableShellExtGlob {
				tks = append(tks, punctQuestionParen)
				continue
			}
			if cfg.allowQuestion {
				tks = append(tks, punctuation('?'))
			} else {
				tks = append(tks, literal('?'))
			}

		case '+':
			// Could be + or +( if shell extglob is enabled
			prev = 0
			if c == '(' && cfg.enableShellExtGlob {
				tks = append(tks, punctPlusParen)
				continue
			}
			tks = append(tks, literal('+'))

		case '@':
			// Could be @ or @( if shell extglob is enabled
			prev = 0
			if c == '(' && cfg.enableShellExtGlob {
				tks = append(tks, punctAtParen)
				continue
			}
			tks = append(tks, literal('@'))

		case '!':
			// Could be ! or !( if shell extglob is enabled
			prev = 0
			if c == '(' && cfg.enableShellExtGlob {
				tks = append(tks, punctBangParen)
				continue
			}
			tks = append(tks, literal('!'))
		}

		switch c {
		case '*': // note prev != '*'
			// It could be a * or ** or *( depending on options.
			switch {
			case (cfg.allowStar && cfg.allowDoubleStar) || cfg.enableShellExtGlob:
				// It could be ** or *(
				prev = '*'

			case cfg.allowStar:
				// It has to be *
				tks = append(tks, punctuation('*'))

			default:
				// * is not allowed to be anything special.
				tks = append(tks, literal('*'))
			}

		case escapeChar:
			if !cfg.allowEscaping {
				tks = append(tks, literal(escapeChar))
				break
			}
			// Next char is escaped.
			prev = escapeChar

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
			if cfg.enableShellExtGlob {
				// It could be ? or ?(
				prev = '?'
				continue
			}
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

		case '+', '@', '!':
			// These could be +(, @(, or !( if shell extglob is enabled
			if cfg.enableShellExtGlob {
				prev = c
			} else {
				tks = append(tks, literal(c))
			}

		case '|', ')':
			// This is the shell extglob special separator and terminator
			if cfg.enableShellExtGlob {
				tks = append(tks, punctuation(c))
			} else {
				tks = append(tks, literal(c))
			}

		default:
			// It's a literal.
			tks = append(tks, literal(c))
		}
	}

	// Unprocessed 'prev' at end of string?
	switch prev {
	case escapeChar:
		tks = append(tks, literal(escapeChar))

	case '*':
		if cfg.allowStar {
			tks = append(tks, punctuation('*'))
		} else {
			tks = append(tks, literal('*'))
		}

	case '?':
		if cfg.allowQuestion {
			tks = append(tks, punctuation('?'))
		} else {
			tks = append(tks, literal('?'))
		}

	case '+', '@', '!':
		tks = append(tks, literal(prev))
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
