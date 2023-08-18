package zzglob

type expression interface {
	// match reports if the rune matches the expression, and whether to
	// keep the current state.
	match(rune) (matched, keep bool)
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

func (e literalExp) match(r rune) (bool, bool) { return rune(e) == r, false }
func (starExp) match(r rune) (bool, bool)      { return r != '/', true }
func (doubleStarExp) match(rune) (bool, bool)  { return true, true }
func (questionExp) match(r rune) (bool, bool)  { return r != '/', false }
