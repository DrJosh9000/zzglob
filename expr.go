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

type state struct {
	Out []edge
}

func (s *state) terminal() bool { return len(s.Out) == 0 }

func matchSegment(set []*state, segment string) []*state {
	var next []*state
	for _, r := range segment {
		if len(set) == 0 {
			return nil
		}
		for _, n := range set {
			for _, e := range n.Out {
				matched, keep := e.Expr.match(r)
				if !matched {
					continue
				}
				next = append(next, e.State)
				if keep {
					next = append(next, n)
				}
			}
		}
		set, next = next, set[:0]
	}
	return set
}

func match(start *state, path string) bool {
	set := matchSegment([]*state{start}, path)
	for _, n := range set {
		if n.terminal() {
			return true
		}
	}
	return false
}
