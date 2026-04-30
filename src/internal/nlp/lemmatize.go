package nlp

import "strings"

type suffixRule struct {
	suffix  string
	minStem int
}

// suffixRules ordered longest-first so the most specific suffix wins.
var suffixRules = []suffixRule{
	{"amiento", 3},
	{"imiento", 3},
	{"amente", 3},
	{"mente", 3},
	{"acion", 3},
	{"icion", 3},
	{"cion", 3},
	{"sion", 3},
	{"ando", 3},
	{"iendo", 3},
	{"aron", 3},
	{"ieron", 3},
	{"izar", 3},
	{"idad", 3},
	{"ado", 3},
	{"ada", 3},
	{"ido", 3},
	{"ida", 3},
	{"oso", 3},
	{"osa", 3},
}

// Lemmatize applies Spanish suffix-stripping rules to a single token.
// Returns the stem if a rule matches and the remaining stem is long enough,
// or the original token otherwise.
func Lemmatize(tok string) string {
	for _, rule := range suffixRules {
		if strings.HasSuffix(tok, rule.suffix) {
			stem := tok[:len(tok)-len(rule.suffix)]
			if len(stem) >= rule.minStem {
				return stem
			}
		}
	}
	return tok
}
