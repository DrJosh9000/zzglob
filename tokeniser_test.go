package zzglob

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTokeniser(t *testing.T) {
	tests := []struct {
		pattern string
		want    *tokens
	}{
		{
			pattern: "ab/cde",
			want: &tokens{
				token('a'),
				token('b'),
				token('/'),
				token('c'),
				token('d'),
				token('e'),
			},
		},
		{
			pattern: "\\ jam\\",
			want: &tokens{
				token(' '),
				token('j'),
				token('a'),
				token('m'),
				token('\\'),
			},
		},
		{
			pattern: "* or ** or \\*? *",
			want: &tokens{
				tokenStar,
				token(' '),
				token('o'),
				token('r'),
				token(' '),
				tokenDoubleStar,
				token(' '),
				token('o'),
				token('r'),
				token(' '),
				token('*'),
				tokenQuestion,
				token(' '),
				tokenStar,
			},
		},
		{
			pattern: "{a,b,c}[d*\\]e]]",
			want: &tokens{
				tokenOpenBrace,
				token('a'),
				tokenComma,
				token('b'),
				tokenComma,
				token('c'),
				tokenCloseBrace,
				tokenOpenBracket,
				token('d'),
				token('*'),
				token(']'),
				token('e'),
				tokenCloseBracket,
				token(']'),
			},
		},
	}

	// Fix the config in case this test is ever run on Windows.
	cfg := defaultParseConfig
	cfg.allowEscaping = true
	cfg.swapSlashes = false

	for _, test := range tests {
		got := tokenise(test.pattern, &cfg)
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Errorf("tokenise(%q) diff (-got +want):\n%s", test.pattern, diff)
		}
	}
}

func TestTokeniser_SwapSlashes(t *testing.T) {
	tests := []struct {
		pattern string
		want    *tokens
	}{
		{
			pattern: "ab/cde",
			want: &tokens{
				token('a'),
				token('b'),
				token('c'),
				token('d'),
				token('e'),
			},
		},
		{
			pattern: "\\ jam\\",
			want: &tokens{
				token('/'),
				token(' '),
				token('j'),
				token('a'),
				token('m'),
				token('/'),
			},
		},
		{
			pattern: "* or ** or \\*? *",
			want: &tokens{
				tokenStar,
				token(' '),
				token('o'),
				token('r'),
				token(' '),
				tokenDoubleStar,
				token(' '),
				token('o'),
				token('r'),
				token(' '),
				token('/'),
				tokenStar,
				tokenQuestion,
				token(' '),
				tokenStar,
			},
		},
		{
			pattern: "{a,b,c}[d*\\]e]]",
			want: &tokens{
				tokenOpenBrace,
				token('a'),
				tokenComma,
				token('b'),
				tokenComma,
				token('c'),
				tokenCloseBrace,
				tokenOpenBracket,
				token('d'),
				token('*'),
				token('\\'),
				tokenCloseBracket,
				token('e'),
				token(']'),
				token(']'),
			},
		},
	}

	// Fix the config in case this test is ever run on !Windows.
	cfg := defaultParseConfig
	cfg.allowEscaping = true
	cfg.swapSlashes = true

	for _, test := range tests {
		got := tokenise(test.pattern, &cfg)
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Errorf("tokenise(%q) diff (-got +want):\n%s", test.pattern, diff)
		}
	}
}
