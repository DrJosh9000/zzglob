package zzglob

// Lexer tokens
type (
	literal     rune // not any of the below
	punctuation rune // *, ** (as ⁑), ?, {, }, [, ], or comma
)

func (literal) tokenTag()     {}
func (punctuation) tokenTag() {}

type token interface{ tokenTag() }

type tokens []token

func tokenise(p string, cfg *parseConfig) *tokens {
	// Most tokens are single runes, so preallocate len(p).
	tks := make(tokens, 0, len(p))

	// Tokenisation state.
	escape := false   // the previous char was \
	star := false     // the previous char was *
	insideCC := false // within a char class

	// Walk through string, producing tokens.
	for _, c := range p {
		// Escaping something?
		if escape {
			// The \ escaped c - c is a literal.
			escape = false
			tks = append(tks, literal(c))
			continue
		}

		// Within a char class? No escaping required, other than ].
		// But the user can optionally escape everything, and must if they want
		// ], so escape is higher priority.
		if insideCC {
			switch c {
			case '\\':
				// Start of escape
				escape = true

			case ']':
				// End of cc
				tks = append(tks, punctuation(']'))
				insideCC = false

			default:
				tks = append(tks, literal(c))
			}
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
				star = true
			} else {
				tks = append(tks, literal('*'))
			}

		case '\\':
			if !cfg.allowEscaping {
				tks = append(tks, literal('\\'))
				break
			}
			// Next char is escaped.
			escape = true

		case '[':
			if !cfg.allowCharClass {
				tks = append(tks, literal('['))
				break
			}
			insideCC = true
			fallthrough

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
		tks = append(tks, literal('\\'))
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
