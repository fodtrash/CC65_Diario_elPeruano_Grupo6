package nlp

import "strings"

var stopwords = map[string]struct{}{
	"a": {}, "al": {}, "algo": {}, "ante": {},
	"como": {}, "con": {},
	"de": {}, "del": {}, "desde": {}, "dos": {},
	"e": {}, "el": {}, "en": {}, "entre": {}, "es": {}, "esta": {}, "este": {},
	"fue": {},
	"ha": {}, "han": {}, "hasta": {}, "hay": {},
	"la": {}, "las": {}, "le": {}, "lo": {}, "los": {},
	"mas": {}, "me": {}, "muy": {},
	"ni": {}, "no": {},
	"o": {},
	"para": {}, "pero": {}, "por": {},
	"que": {}, "quien": {},
	"se": {}, "ser": {}, "si": {}, "sin": {}, "sobre": {}, "son": {}, "su": {}, "sus": {},
	"tambien": {}, "todo": {},
	"un": {}, "una": {}, "uno": {},
	"y": {}, "ya": {},
}

// Tokenize splits a cleaned string by whitespace, removes Spanish stopwords
// and tokens shorter than 2 characters.
func Tokenize(s string) []string {
	words := strings.Fields(s)
	result := make([]string, 0, len(words)/2)
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if _, ok := stopwords[w]; ok {
			continue
		}
		result = append(result, w)
	}
	return result
}
