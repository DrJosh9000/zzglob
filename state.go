package zzglob

import "strings"

type state struct {
	Out []edge
}

type edge struct {
	Expr  expression
	State *state
}

func (s *state) terminal() bool { return len(s.Out) == 0 }

func matchSegment(start map[*state]struct{}, segment string) map[*state]struct{} {
	a := make(map[*state]struct{}, len(start))
	b := make(map[*state]struct{}, len(start))
	for n := range start {
		a[n] = struct{}{}
	}

	for _, r := range segment {
		if len(a) == 0 {
			return nil
		}
		for n := range a {
			for _, e := range n.Out {
				matched, keep := e.Expr.match(r)
				if !matched {
					continue
				}
				b[e.State] = struct{}{}
				if keep {
					b[n] = struct{}{}
				}
			}
		}
		a, b = b, a
		clear(b)
	}
	return a
}

func match(root string, start *state, path string) bool {
	rem, ok := strings.CutPrefix(path, root)
	if !ok {
		return false
	}
	set := matchSegment(singleton(start), rem)
	for n := range set {
		if n.terminal() {
			return true
		}
	}
	return false
}

// singleton wraps a single value into a map used as a set.
func singleton[K comparable](k K) map[K]struct{} {
	return map[K]struct{}{k: {}}
}
