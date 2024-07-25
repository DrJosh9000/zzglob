package zzglob

// state represents a possible state of a state machine.
type state struct {
	// Out contains all possible transitions out of this state.
	Out []edge

	// Accept is whether the state is a fully-matched state.
	Accept bool
}

// stateSet represents a set of possible machine states.
type stateSet map[*state]struct{}

// edge represents a state transition inside the state machine.
type edge struct {
	// Expr tests a rune; if the expression passes, the edge can be followed.
	// If Expr is nil, then the edge should be followed before processing
	// the next rune.
	Expr expression

	// State is the machine state that the machine transitions into when Expr
	// is satisfied.
	State *state
}

// singleton wraps a single value in a set.
func singleton(s *state) stateSet { return stateSet{s: {}} }

// matchSegment progresses an initial set of states, one rune from the segment
// at a time.
func matchSegment(initial stateSet, segment string) stateSet {
	a := make(stateSet, len(initial))
	b := make(stateSet, len(initial))
	for n := range initial {
		a[n] = struct{}{}
	}
	transitiveClosure(a)

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
					// The queue should already contain e.State because of
					// transitiveClosure.
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
		transitiveClosure(a)
	}
	return a
}

// transitiveClosure adds any states reachable through edges with nil Expr to
// the same set.
func transitiveClosure(states stateSet) {
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
