package zzglob

import (
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		pattern, path string
		want          bool
	}{
		{"a/b", "a/b", true},
		{"a/b", "a/b/", false},
		{"a*b", "acccccb", true},
		{"a*b", "abc", false},
		{"a*b", "a/b", false},
		{"a/{b,c}/d", "a/c/d", true},
		{"a/{b,c}/d", "a/w/d", false},
		{"a/[bc]/d", "a/b/d", true},
		{"a/[bc]/d", "a/x/d", false},
		{"a/[bc]/d", "b/c/d", false},
		{"a?b", "acb", true},
		{"a?b", "accb", false},
		{"a**b", "acb", true},
		{"a**b", "acccb", true},
		{"a**b", "a/b", true},
		{"a**b", "a/c/b", true},
		{"a**b", "a/b/c", false},
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/c/b", true},
		{"a/**/b", "a/b/c", false},
		{"a/**/b", "a/c/d/e/f/b", true},
		{"*", "a", true},
		{"*", "abcde", true},
		{"*", "abc/", false},
		{"**", "a", true},
		{"**", "abcde", true},
		{"**", "abc/", true},
		{"{a,b*}", "a", true},
		{"{a,b*}", "b", true},
		{"{a,b*}", "ac", false},
		{"{a,b*}", "bc", true},
		{"{a,b*}", "bcc/cc", false},
		{"{a,b**}", "a", true},
		{"{a,b**}", "b", true},
		{"{a,b**}", "ac", false},
		{"{a,b**}", "bc", true},
		{"{a,b**}", "bcc/cc", true},
	}

	for _, test := range tests {
		p, err := Parse(test.pattern)
		if err != nil {
			t.Fatalf("Parse(%q) error = %v", test.pattern, err)
		}

		if got, want := p.Match(test.path), test.want; got != want {
			t.Errorf("(%q).Match(%q) = %v, want %v", test.pattern, test.path, got, want)
			//p.WriteDot(os.Stderr)
		}
	}
}