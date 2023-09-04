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
		if n.Terminal {
			return true
		}
	}
	return false
}

// transitive adds any states reachable through edges with nil expression.
func transitive(states map[*state]struct{}) {
	q := make([]*state, 0, len(states))
	for n := range states {
		q = append(q, n)
	}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]

		for _, e := range n.Out {
			if e.Expr != nil {
				continue
			}
			if _, seen := states[e.State]; seen {
				continue
			}
			states[e.State] = struct{}{}
			q = append(q, e.State)
		}
	}
}

func matchSegment(initial map[*state]struct{}, segment string) map[*state]struct{} {
	a := make(map[*state]struct{}, len(initial))
	b := make(map[*state]struct{}, len(initial))
	for n := range initial {
		a[n] = struct{}{}
	}
	transitive(a)

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
					// The queue should already contain e.State.
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
		transitive(a)
	}
	return a
}
