package zzglob

import (
	"path/filepath"
)

var defaultParseConfig = parseConfig{
	allowEscaping:    filepath.Separator == '/',
	allowQuestion:    true,
	allowStar:        true,
	allowDoubleStar:  true,
	allowAlternation: true,
	allowCharClass:   true,
	swapSlashes:      filepath.Separator != '/',
	expandTilde:      true,
}

type parseConfig struct {
	allowEscaping    bool
	allowQuestion    bool
	allowStar        bool
	allowDoubleStar  bool
	allowAlternation bool
	allowCharClass   bool
	swapSlashes      bool
	expandTilde      bool
}

// ParseOption functions optionally alter how patterns are parsed.
type ParseOption = func(*parseConfig)

// AllowEscaping changes how the escape character (usually backslash - see
// WithSwapSlashes), is parsed. If disabled, it is treated as a literal which
// does not escape the next character.
// By default, AllowEscaping is disabled if filepath.Separator is not /
// (i.e. on Windows) and enabled otherwise.
func AllowEscaping(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowEscaping = enable
	}
}

// AllowQuestion changes how ? is parsed. If disabled, ? is treated as a
// literal. Enabled by default.
func AllowQuestion(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowQuestion = enable
	}
}

// AllowStar changes how * is parsed. If disabled, * is treated as a literal.
// Enabled by default.
func AllowStar(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowStar = enable
	}
}

// AllowDoubleStar changes how ** is parsed, and applies only if AllowStar is
// enabled (the default). If disabled, ** is treated as two consecutive
// instances of * (equivalent to a single *). Enabled by default.
func AllowDoubleStar(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowDoubleStar = enable
	}
}

// AllowAlternation changes how { } are parsed. If enabled, { and } delimit
// alternations. If disabled, { and } are treated as literals.
// Enabled by default.
func AllowAlternation(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowAlternation = enable
	}
}

// AllowCharClass changes how [ ] are parsed. If enabled, [ and ] denote
// character classes. If disabled, [ and ] are treated as literals.
// Enabled by default.
func AllowCharClass(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowCharClass = enable
	}
}

// ExpandTilde changes how ~ is parsed. If enabled, ~ is expanded to the current
// user's home directory. If disabled, ~ is treated as a literal.
// Enabled by default.
func ExpandTilde(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.expandTilde = enable
	}
}

// WithSwapSlashes changes how \ and / are interpreted. If enabled, / becomes the
// escape character (which can be disabled with AllowEscaping), and \ becomes
// the path separator (typical on Windows). Note that after parsing, the pattern
// internally uses / to represent the path separator (which is consistent with
// io/fs). To receive the correct slashes from Match or Glob, be sure to use
// the TranslateSlashes option.
// By default, WithSwapSlashes is enabled if filepath.Separator is not /
// (i.e. on Windows) and disabled otherwise.
func WithSwapSlashes(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.swapSlashes = enable
	}
}
