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
		if n.Accept {
			return true
		}
	}
	return false
}
