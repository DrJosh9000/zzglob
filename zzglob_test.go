package zzglob

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTokeniser(t *testing.T) {
	tests := []struct {
		pattern string
		want    []component
	}{
		{
			pattern: "ab/cde",
			want: []component{
				literal('a'),
				literal('b'),
				pathSeparator{},
				literal('c'),
				literal('d'),
				literal('e'),
			},
		},
		{
			pattern: "\\ jam\\",
			want: []component{
				literal(' '),
				literal('j'),
				literal('a'),
				literal('m'),
				literal('\\'),
			},
		},
		{
			pattern: "* or ** or \\*? *",
			want: []component{
				star{},
				literal(' '),
				literal('o'),
				literal('r'),
				literal(' '),
				doubleStar{},
				literal(' '),
				literal('o'),
				literal('r'),
				literal(' '),
				literal('*'),
				question{},
				literal(' '),
				star{},
			},
		},
		{
			pattern: "{a,b,c}[de]",
			want: []component{
				openBrace{},
				literal('a'),
				comma{},
				literal('b'),
				comma{},
				literal('c'),
				closeBrace{},
				openBracket{},
				literal('d'),
				literal('e'),
				closeBracket{},
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
