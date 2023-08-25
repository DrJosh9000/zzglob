package zzglob

import (
	"fmt"
	"io"
	"strings"
)

type state struct {
	Out []edge
}

type edge struct {
	Expr  expression
	State *state
}

func (s *state) terminal() bool { return len(s.Out) == 0 }

func writeDot(w io.Writer, start *state) error {
	seen := make(map[*state]bool)
	q := []*state{start}
	for len(q) > 0 {
		s := q[0]
		q = q[1:]

		if seen[s] {
			continue
		}
		seen[s] = true

		if _, err := fmt.Fprintf(w, "state_%p [label=\"\"];\n", s); err != nil {
			return err
		}
		for _, e := range s.Out {
			if _, err := fmt.Fprintf(w, "state_%p -> state_%p [label=\"%v\"];\n", s, e.State, e.Expr); err != nil {
				return err
			}
			if seen[e.State] {
				continue
			}
			q = append(q, e.State)
		}
	}
	return nil
}

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
