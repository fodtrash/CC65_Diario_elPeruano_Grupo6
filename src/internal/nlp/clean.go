package nlp

import (
	"regexp"
	"strings"
)

var accentReplacer = strings.NewReplacer(
	"\u00e1", "a", // á
	"\u00e9", "e", // é
	"\u00ed", "i", // í
	"\u00f3", "o", // ó
	"\u00fa", "u", // ú
	"\u00f1", "n", // ñ
	"\u00fc", "u", // ü
	"\u00c1", "a", // Á
	"\u00c9", "e", // É
	"\u00cd", "i", // Í
	"\u00d3", "o", // Ó
	"\u00da", "u", // Ú
	"\u00d1", "n", // Ñ
	"\u00dc", "u", // Ü
)

var reNonAlnum = regexp.MustCompile(`[^a-z0-9 ]`)
var reMultiSpace = regexp.MustCompile(`\s{2,}`)

// Clean normalizes a raw Spanish text string: lowercase, remove accents,
// strip non-alphanumeric characters (keeping spaces), collapse multiple
// spaces, and trim.
func Clean(s string) string {
	s = strings.ToLower(s)
	s = accentReplacer.Replace(s)
	s = reNonAlnum.ReplaceAllString(s, " ")
	s = reMultiSpace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}
