package zzglob

type expression interface {
	match(rune) (match, split bool)
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
	Expr expression
	Node *node
}

type node struct {
	Out []edge
}

func (n *node) terminal() bool { return len(n.Out) == 0 }

func match(start *node, path string) bool {
	nodes := []*node{start}
	var next []*node
	for _, r := range path {
		if len(nodes) == 0 {
			return false
		}
		for _, n := range nodes {
			for _, e := range n.Out {
				m, s := e.Expr.match(r)
				if m {
					next = append(next, e.Node)
					if s {
						next = append(next, n)
					}
				}
			}
		}
		nodes, next = next, nodes[:0]
	}
	for _, n := range nodes {
		if n.terminal() {
			return true
		}
	}
	return false
}
