package zzglob

import (
	"strings"
)

type state struct {
	Out []edge
}

type edge struct {
	Expr  expression
	State *state
}

// Match reports if the path matches the pattern.
func (p *Pattern) Match(path string) bool {
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
		for n := range a {
			for _, e := range n.Out {
				matched := e.Expr.match(r)
				if !matched {
					continue
				}
				b[e.State] = struct{}{}
			}
		}
		a, b = b, a
		clear(b)
	}
	return a
}

// singleton wraps a single value into a map used as a set.
func singleton[K comparable](k K) map[K]struct{} {
	return map[K]struct{}{k: {}}
}

func (s *state) terminal() bool { return len(s.Out) == 0 }
