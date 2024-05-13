package zzglob

import "slices"

type expression interface {
	// match reports if the rune matches the expression
	match(rune) bool
}

// Expressions
type (
	// Literal expression
	// Matches exactly this literal (including the path separator, /)
	literalExp rune

	// Star expression
	// * matches like regexp [^/]*
	starExp struct{}

	// Double-star expression
	// ** matches like regexp .*
	doubleStarExp struct{}

	// Wildcard character expression
	// ? matches like regexp [^/]
	questionExp struct{}

	// Negated character class expression
	// (Non-negated char classes are implemented like alternations: multiple
	// out-edges from a state. Negating the negation in order to use the same
	// representation won't work here: it would consist of a vast number of
	// potentially matching runes.)
	// [^...] matches like regexp [^...]
	// The value of a negatedCCExp contains all the runes that do *not* match.
	negatedCCExp []rune

	// The single-rune version of negatedCCExp
	negatedLiteralExp rune
)

func (e literalExp) match(r rune) bool { return rune(e) == r }
func (starExp) match(r rune) bool      { return r != '/' }
func (doubleStarExp) match(rune) bool  { return true }
func (questionExp) match(r rune) bool  { return r != '/' }

func (e negatedCCExp) match(r rune) bool {
	// Unsubstantiated claim: negated classes usually contain few runes,
	// so a linear search is probably acceptably fast, a binary search is
	// probably very fast, and a map lookup might not be worth the work.
	_, found := slices.BinarySearch(e, r)
	return !found
}

func (e negatedLiteralExp) match(r rune) bool { return rune(e) != r }

func (e literalExp) String() string        { return string(e) }
func (starExp) String() string             { return "*" }
func (doubleStarExp) String() string       { return "**" }
func (questionExp) String() string         { return "?" }
func (e negatedCCExp) String() string      { return "[^" + string(e) + "]" }
func (e negatedLiteralExp) String() string { return "[^" + string(e) + "]" }
