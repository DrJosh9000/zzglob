package zzglob

type expression interface {
	match(s string) (match, done bool)
}

// Expressions
type (
	// Matches exactly this literal
	literalExp rune

	// A sequence matches expressions in sequence.
	sequenceExp []expression

	// Matches / (the path separator)
	pathSepExp struct{}

	// * matches like [^/]*
	starExp struct{}

	// ** matches like .*
	doubleStarExp struct{}

	// ? matches like .
	questionExp struct{}

	// Matches a set of runes.
	charClassExp map[rune]struct{}

	// Matches any of the expressions it contains.
	alternationExp []expression
)

func (e literalExp) match(s string) (match, done bool) {
	if s == string(e) {
		return true, true
	}
	for _, r := range s {
		return r == rune(e), false
	}
	return false, false
}

func (e sequenceExp) match(s string) (match, done bool) {

}
