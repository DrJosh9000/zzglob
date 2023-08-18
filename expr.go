package zzglob

type expression interface {
	// match reports if the rune matches the expression, and whether to
	// keep the current state.
	match(rune) (matched, keep bool)
}

// Expressions
type (
	// Matches exactly this literal
	literalExp rune

	// Matches / (the path separator)
	pathSepExp struct{}

	// * matches like [^/]*
	starExp struct{}

	// ** matches like .*
	doubleStarExp struct{}

	// ? matches like .
	questionExp struct{}
)

func (e literalExp) match(r rune) (bool, bool) { return rune(e) == r, false }
func (pathSepExp) match(r rune) (bool, bool)   { return r == '/', false }
func (starExp) match(r rune) (bool, bool)      { return r != '/', true }
func (doubleStarExp) match(rune) (bool, bool)  { return true, true }
func (questionExp) match(r rune) (bool, bool)  { return r != '/', false }

type edge struct {
	Expr  expression
	State *state
}
