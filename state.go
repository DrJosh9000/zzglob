package zzglob

type state struct {
	Out []edge
}

type edge struct {
	Expr  expression
	State *state
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
