package zzglob

import "strings"

// Match reports if the path matches the pattern.
func (p *Pattern) Match(path string) bool {
	if p.initial == nil {
		// no state machine, only root
		return path == p.root
	}

	rem, ok := strings.CutPrefix(path, p.root)
	if !ok {
		return false
	}
	set := matchSegment(singleton(p.initial), rem)
	for n := range set {
		if n.terminal() {
			return true
		}
	}
	return false
}

func matchSegment(initial map[*state]struct{}, segment string) map[*state]struct{} {
	a := make(map[*state]struct{}, len(initial))
	b := make(map[*state]struct{}, len(initial))
	for n := range initial {
		a[n] = struct{}{}
	}

	for _, r := range segment {
		if len(a) == 0 {
			return nil
		}
		for len(a) > 0 {
			// Treating a as a "queue", pop one state (n).
			var n *state
			for x := range a {
				n = x
				break
			}
			delete(a, n)
			for _, e := range n.Out {
				if e.Expr == nil {
					// Nil expression means this edge should be followed
					// right away.
					a[e.State] = struct{}{}
					continue
				}
				matched := e.Expr.match(r)
				if !matched {
					continue
				}
				b[e.State] = struct{}{}
			}
		}
		a, b = b, a
	}
	return a
}
