package zzglob

import (
	"strings"
)

// Lexer tokens
// A non-negative token is a literal (it means the rune that it is).
// A negative token has special meaning (it's a token).
type token rune

type tokens []token

// Special tokens.
const (
	tokenStar         token = -'*' // *
	tokenQuestion     token = -'?' // ?
	tokenOpenBrace    token = -'{' // {
	tokenCloseBrace   token = -'}' // }
	tokenOpenBracket  token = -'[' // [
	tokenCloseBracket token = -']' // ]
	tokenTilde        token = -'~' // ~
	tokenComma        token = -',' // ,
	tokenDoubleStar   token = -128 // **
	tokenBracketCaret token = -129 // [^
)

func (t token) String() string {
	switch t {
	case tokenStar:
		return "*"
	case tokenQuestion:
		return "?"
	case tokenOpenBrace:
		return "{"
	case tokenCloseBrace:
		return "}"
	case tokenOpenBracket:
		return "["
	case tokenCloseBracket:
		return "]"
	case tokenTilde:
		return "~"
	case tokenComma:
		return ","
	case tokenDoubleStar:
		return "**"
	case tokenBracketCaret:
		return "[^"
	}
	return string(rune(t))
}

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
	insideCC := false // within a char class

	// Walk through string, producing tokens.
	for _, c := range p {
		// Escaping something?
		if prev == escapeChar {
			// The escapeChar escaped c, so c is a literal.
			prev = 0
			tks = append(tks, token(c))
			continue
		}

		// Possible double-rune token?
		switch prev {
		case '*':
			// Could be * or ** depending on options
			prev = 0
			if c == '*' && cfg.allowDoubleStar {
				tks = append(tks, tokenDoubleStar)
				continue
			}

			// The previous char was * with possible extra meaning, but c
			// doesn't make it anything special.
			// Emit *, then process c normally.
			if cfg.allowStar {
				tks = append(tks, tokenStar)
			} else {
				tks = append(tks, token('*'))
			}

		case '[':
			prev = 0
			insideCC = true
			if c == '^' {
				tks = append(tks, tokenBracketCaret)
				continue
			} else {
				tks = append(tks, tokenOpenBracket)
			}
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
					tks = append(tks, token(escapeChar))
				}

			case ']':
				// End of cc
				tks = append(tks, tokenCloseBracket)
				insideCC = false

			case '^':
				// Negated char class only if ^ is first char inside [ ]
				// That's handled by prev switch
				tks = append(tks, token('^'))

			default:
				tks = append(tks, token(c))
			}

			continue
		}

		switch c {
		case '*': // note prev != '*'
			// It could be a * or ** depending on options.
			switch {
			case cfg.allowStar && cfg.allowDoubleStar:
				// It could be **
				prev = '*'

			case cfg.allowStar:
				// It has to be *
				tks = append(tks, tokenStar)

			default:
				// * is not allowed to be anything special.
				tks = append(tks, token('*'))
			}

		case escapeChar:
			if !cfg.allowEscaping {
				tks = append(tks, token(escapeChar))
				break
			}
			// Next char is escaped.
			prev = escapeChar

		case pathSep:
			// Always represent the path separator with / for consistency
			// with io/fs.
			tks = append(tks, token('/'))

		case '~':
			if cfg.expandTilde {
				tks = append(tks, tokenTilde)
			} else {
				tks = append(tks, token('~'))
			}

		case '[':
			if cfg.allowCharClass {
				// It could be [ or [^
				prev = '['
			} else {
				tks = append(tks, token('['))
			}

		case '?':
			if cfg.allowQuestion {
				tks = append(tks, tokenQuestion)
			} else {
				tks = append(tks, token('?'))
			}

		case ']':
			// We only get here if insideCC is false...
			tks = append(tks, token(']'))

		case '{', '}', ',':
			if cfg.allowAlternation {
				switch c {
				case '{':
					tks = append(tks, tokenOpenBrace)
				case '}':
					tks = append(tks, tokenCloseBrace)
				case ',':
					tks = append(tks, tokenComma)
				}
			} else {
				tks = append(tks, token(c))
			}

		default:
			// It's a literal.
			tks = append(tks, token(c))
		}
	}

	// Unprocessed 'prev' at end of string?
	switch prev {
	case '*':
		if cfg.allowStar {
			tks = append(tks, tokenStar)
		} else {
			tks = append(tks, token('*'))
		}

	case 0:
		// Nothing unprocessed

	default: // escapeChar or [
		tks = append(tks, token(prev))
	}

	return &tks
}

// next uses a pointer to a slice as a consuming reader.
func (r *tokens) next() (token, bool) {
	if r == nil || len(*r) == 0 {
		return 0, false
	}
	defer func() { *r = (*r)[1:] }()
	return (*r)[0], true
}

// allLiteral returns a string consisting of all tokens runic equivalents.
// r must consist solely of literals, otherwise it returns the empty string.
func (r tokens) allLiteral() string {
	b := strings.Builder{}
	b.Grow(len(r))
	for _, t := range r {
		if t < 0 {
			return ""
		}
		b.WriteRune(rune(t))
	}
	return b.String()
}
