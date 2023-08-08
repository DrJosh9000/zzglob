package zzglob

type expression interface {
	match(rune) (matched, proceed bool)
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

func (e literalExp) match(r rune) (bool, bool) { return rune(e) == r, true }
func (pathSepExp) match(r rune) (bool, bool)   { return r == '/', true }
func (starExp) match(r rune) (bool, bool)      { return r != '/', false }
func (doubleStarExp) match(rune) (bool, bool)  { return true, false }
func (questionExp) match(r rune) (bool, bool)  { return r != '/', true }

type edge struct {
	Expr expression
	Node *node
}

type node struct {
	Out []edge
}
