package zzglob

import (
	"os/user"
	"path/filepath"
	"strings"
)

// preprocess preprocesses the token sequence in the following ways.
// Because ** should match zero path segments:
// - prefix **/ becomes {,**/}
// - /**/ becomes /{,**/}
// Because ~ means homedir:
// - Prefix ~/ becomes homedir/ (but only if user.Current() succeeds.)
//
// TODO: Arbitrary local user homedirs?
func preprocess(in tokens) tokens {
	type replacement struct {
		find, sub tokens
	}
	prefixSubs := []replacement{
		{
			// Prefix **/ becomes {,**/}
			find: tokens{tokenDoubleStar, token('/')},
			sub: tokens{
				tokenOpenBrace, tokenComma, tokenDoubleStar,
				token('/'), tokenCloseBrace,
			},
		},
	}

	if hd := homeDir(); len(hd) > 0 {
		prefixSubs = append(prefixSubs, replacement{
			// Prefix ~/ becomes homedir/
			find: tokens{tokenTilde, token('/')},
			sub:  hd,
		})
	}

	for _, ps := range prefixSubs {
		in = replacePrefix(in, ps.find, ps.sub)
	}

	allSubs := []replacement{
		{
			// /**/ becomes /{,**/}
			find: tokens{
				token('/'), tokenDoubleStar, token('/'),
			},
			sub: tokens{
				token('/'), tokenOpenBrace, tokenComma,
				tokenDoubleStar, token('/'), tokenCloseBrace,
			},
		},
	}

	for _, as := range allSubs {
		in = replaceAll(in, as.find, as.sub)
	}

	return in
}

// homeDir returns the current user's homedir as a literal token sequence.
func homeDir() tokens {
	u, err := user.Current()
	if err != nil {
		// Oh well, no homedir for you.
		return nil
	}
	homeDir := filepath.ToSlash(u.HomeDir)
	hd := make(tokens, 0, len(u.HomeDir))
	for _, r := range homeDir {
		hd = append(hd, token(r))
	}
	if !strings.HasSuffix(homeDir, "/") {
		hd = append(hd, token('/'))
	}
	return hd
}

// hasPrefix reports whether in has the prefix.
func hasPrefix(in, prefix tokens) bool {
	if len(in) < len(prefix) {
		return false
	}
	for i, t := range prefix {
		if t != in[i] {
			return false
		}
	}
	return true
}

// replacePrefix replaces the prefix with sub, if in has prefix as a prefix.
func replacePrefix(in, prefix, sub tokens) tokens {
	if !hasPrefix(in, prefix) {
		return in
	}
	return append(sub, in[len(prefix):]...)
}

// replaceAll replaces all non-overlapping instances of toFind with sub.
func replaceAll(in, toFind, sub tokens) tokens {
	out := make([]token, 0, len(in))
	next := 0
	for _, t := range in {
		if t == toFind[next] {
			next++
			if next == len(toFind) {
				out = append(out, sub...)
				next = 0
			}
		} else {
			if next != 0 {
				out = append(out, toFind[:next]...)
				next = 0
			}
			out = append(out, t)
		}
	}
	if next != 0 {
		out = append(out, toFind[:next]...)
	}
	return out
}

// findRoot returns the longest prefix consisting of literals, up to (including)
// the final path separator. tks is trimmed to be the remainder of the pattern.
func findRoot(tks *tokens) string {
	var root []rune
	lastSlash := -1
	for i, t := range *tks {
		if t < 0 {
			break
		}
		if t == '/' {
			lastSlash = i
		}
		root = append(root, rune(t))
	}
	if lastSlash < 0 {
		// No slash, no root.
		return ""
	}
	*tks = (*tks)[lastSlash+1:]
	return string(root[:lastSlash+1])
}
