package zzglob

type expression interface {
	// match reports if the rune matches the expression
	match(rune) bool
}

// Expressions
type (
	// Matches exactly this literal (including the path separator, /)
	literalExp rune

	// * matches like [^/]*
	starExp struct{}

	// ** matches like .*
	doubleStarExp struct{}

	// ? matches like [^/]
	questionExp struct{}
)

func (e literalExp) match(r rune) bool { return rune(e) == r }
func (starExp) match(r rune) bool      { return r != '/' }
func (doubleStarExp) match(rune) bool  { return true }
func (questionExp) match(r rune) bool  { return r != '/' }

func (e literalExp) String() string  { return string(e) }
func (starExp) String() string       { return "*" }
func (doubleStarExp) String() string { return "**" }
func (questionExp) String() string   { return "?" }
