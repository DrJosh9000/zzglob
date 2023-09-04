package zzglob

// ParseOption functions optionally alter how patterns are parsed.
type ParseOption = func(*parseConfig)

type parseConfig struct {
	allowEscaping    bool
	allowQuestion    bool
	allowStar        bool
	allowDoubleStar  bool
	allowAlternation bool
	allowCharClass   bool
}

// AllowEscaping changes how \ is parsed. If disabled, \ is treated as a
// literal and does not escape the next character.
func AllowEscaping(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowEscaping = enable
	}
}

// AllowQuestion changes how ? is parsed. If disabled, ? is treated as a
// literal.
func AllowQuestion(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowQuestion = enable
	}
}

// AllowStar changes how * is parsed. If disabled, * is treated as a literal.
func AllowStar(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowStar = enable
	}
}

// AllowDoubleStar changes how ** is parsed, and applies only if AllowStar is
// enabled. If disabled, ** is treated as two consecutive instances of *
// (equivalent to a single *).
func AllowDoubleStar(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowDoubleStar = enable
	}
}

// AllowAlternation changes how { } are parsed. If disabled, { and } are treated
// as literals.
func AllowAlternation(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowAlternation = enable
	}
}

// AllowCharClass changes how [ ] are parsed. If disabled, [ and ] are treated
// as literals.
func AllowCharClass(enable bool) ParseOption {
	return func(o *parseConfig) {
		o.allowCharClass = enable
	}
}
