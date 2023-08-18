package zzglob

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTokeniser(t *testing.T) {
	tests := []struct {
		pattern string
		want    tokens
	}{
		{
			pattern: "ab/cde",
			want: tokens{
				literal('a'),
				literal('b'),
				literal('/'),
				literal('c'),
				literal('d'),
				literal('e'),
			},
		},
		{
			pattern: "\\ jam\\",
			want: tokens{
				literal(' '),
				literal('j'),
				literal('a'),
				literal('m'),
				literal('\\'),
			},
		},
		{
			pattern: "* or ** or \\*? *",
			want: tokens{
				punctuation('*'),
				literal(' '),
				literal('o'),
				literal('r'),
				literal(' '),
				punctuation('⁑'),
				literal(' '),
				literal('o'),
				literal('r'),
				literal(' '),
				literal('*'),
				punctuation('?'),
				literal(' '),
				punctuation('*'),
			},
		},
		{
			pattern: "{a,b,c}[d*\\]e]",
			want: tokens{
				punctuation('{'),
				literal('a'),
				punctuation(','),
				literal('b'),
				punctuation(','),
				literal('c'),
				punctuation('}'),
				punctuation('['),
				literal('d'),
				literal('*'),
				literal(']'),
				literal('e'),
				punctuation(']'),
			},
		},
	}

	for _, test := range tests {
		got := tokenise(test.pattern)
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Errorf("tokenise(%q) diff (-got +want):\n%s", test.pattern, diff)
		}
	}
}
