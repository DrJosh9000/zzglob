// Package zzglob implements a file path walker.
package zzglob

import (
	"errors"
	"io/fs"
	"path/filepath"
)

// Glob globs for files matching the pattern.
func Glob(pattern string, f fs.WalkDirFunc) error {
	root, _, err := parse(pattern)
	if err != nil {
		return err
	}

	// Roots holds roots to walk.
	// New roots are added in order to traverse symlinks.
	roots := []string{root}
	for len(roots) > 0 {
		root := roots[0]
		roots = roots[1:]

		if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// if pattern.match(path) {
			// 	if err := f(path, d, err); err != nil {
			// 		return err
			// 	}
			// }
			// if pattern.partialMatch(path) {
			// 	// something like that
			// }

			return nil

		}); err != nil {
			return err
		}
	}
	return nil
}

func tokenise(p string) []component {
	// Most tokens are single runes.
	tokens := make([]component, 0, len(p))
	emit := func(t component) { tokens = append(tokens, t) }

	// Tokenisation state.
	esc := false // the previous char was \
	str := false // the previous char was *

	// Walk through string producing tokens.
	for _, c := range p {
		// Escaping something?
		if esc {
			// The \ escaped c, i.e. c is literal.
			esc = false
			emit(literal(c))
			continue
		}

		// Wishing upon a star?
		if str {
			str = false
			if c == '*' {
				// Double star.
				emit(doubleStar{})
				continue
			}

			// The previous char was a star, but this one isn't.
			// Emit *, then process c normally.
			emit(star{})
		}

		switch c {
		case '*': // previous char is not *
			// It could be a star or double star.
			str = true

		case '\\':
			// Next char is escaped.
			esc = true

		case '/':
			emit(pathSeparator{})

		case '?':
			emit(question{})

		case '[':
			emit(openBracket{})

		case ']':
			emit(closeBracket{})

		case '{':
			emit(openBrace{})

		case '}':
			emit(closeBrace{})

		case ',':
			emit(comma{})

		default:
			// It's a literal.
			emit(literal(c))
		}
	}

	// Escape or star at end of string?
	if esc {
		emit(literal('\\'))
	}
	if str {
		emit(star{})
	}

	return tokens
}

func parse(p string) (string, pattern, error) {
	//tokens := tokenise(p)

	return "", pattern{}, errors.New("unimplemented")
}

type component interface{}

// Lexical elements
type (
	// Any rune that tokenised as itself, or was after \
	literal rune

	// Lexical elements with meaning
	pathSeparator struct{} // \
	star          struct{} // *
	doubleStar    struct{} // **
	question      struct{} // ?
	openBrace     struct{} // {
	closeBrace    struct{} // }
	openBracket   struct{} // [
	closeBracket  struct{} // ]
	comma         struct{} // ,
)

// Semantic elements
type (
	// A pattern is one or more components in order.
	pattern []component

	// A char class matches a set of runes.
	charClass map[rune]struct{}

	// An alternation matches any of the patterns it contains.
	alternation []pattern
)
