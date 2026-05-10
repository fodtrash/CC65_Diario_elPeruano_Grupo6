// pipeline.go — Pipeline secuencial de preprocesamiento NLP
//
// Implementa las tres etapas del modelo Promela de forma estrictamente
// secuencial: Lectura → Tokenización BPE → Lematización.
// Sin goroutines ni canales; cada documento se procesa por completo
// antes de pasar al siguiente. Esta versión es la línea base para
// el análisis de speedup y escalabilidad frente a la versión concurrente.
//
// El estado global se actualiza de forma directa, sin mutex ni canales,
// para reflejar un diseño estrictamente secuencial.
package main

import (
	"strings"
	"unicode"
)

// ── Estado global compartido ─────────────────────────────────────────────────
// Equivale a las variables globales del modelo Promela.

// PipelineState agrupa los contadores globales del pipeline.
//
//	Refleja el estado compartido del modelo Promela:
//	docs_procesados, docs_reales, docs_sinteticos, tokens_globales
type PipelineState struct {
	DocsProcessados int
	DocsReales      int
	DocsSinteticos  int
	TokensGlobales  int
}

// ── Stopwords del español ───────────────────────

var stopwordsES = map[string]bool{
	"de": true, "la": true, "que": true, "el": true, "en": true,
	"y": true, "a": true, "los": true, "del": true, "se": true,
	"las": true, "por": true, "un": true, "para": true, "con": true,
	"una": true, "su": true, "al": true, "es": true, "lo": true,
	"como": true, "más": true, "pero": true, "sus": true, "le": true,
	"ya": true, "o": true, "este": true, "si": true, "porque": true,
	"esta": true, "entre": true, "cuando": true, "muy": true, "sin": true,
	"sobre": true, "ser": true, "tiene": true, "también": true, "me": true,
	"hasta": true, "hay": true, "donde": true, "han": true, "quien": true,
	"están": true, "estado": true, "desde": true, "todo": true, "nos": true,
}

// ── Etapa 1: Tokenización BPE (simplificada) ─────────────────────────────────

// tokenize aplica una tokenización BPE simplificada sobre la sumilla.
// Proceso:
//  1. Normaliza unicode y elimina puntuación lateral
//  2. Divide en palabras (pre-tokenización por espacios)
//  3. Aplica merge BPE simulado: si un token supera maxTokenLen,
//     se divide en bigramas de caracteres
func tokenize(doc Document) TokenizedDoc {
	const maxTokenLen = 8

	words := strings.FieldsFunc(doc.Sumilla, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == ';' || r == ':' ||
			r == '(' || r == ')' || r == '"' || r == '\''
	})

	tokens := make([]string, 0, len(words)*2)
	for _, w := range words {
		// Limpia puntuación final (punto, guion)
		w = strings.TrimRight(w, ".-")
		w = strings.TrimLeft(w, "-")
		if w == "" {
			continue
		}
		// BPE merge: tokens largos → bigramas de caracteres
		if len([]rune(w)) > maxTokenLen {
			runes := []rune(w)
			for i := 0; i < len(runes)-1; i += 2 {
				if i+1 < len(runes) {
					tokens = append(tokens, string(runes[i:i+2]))
				} else {
					tokens = append(tokens, string(runes[i]))
				}
			}
		} else {
			tokens = append(tokens, w)
		}
	}
	return TokenizedDoc{Document: doc, Tokens: tokens}
}

// ── Etapa 2: Lematización y normalización ────────────────────────────────────

// lemmatize aplica lematización y eliminación de stopwords.
// Proceso:
//  1. Lowercase de todos los tokens
//  2. Elimina stopwords del español
//  3. Reducción de sufijos (lematización heurística)
func lemmatize(tdoc TokenizedDoc) LematizedDoc {
	lemmas := make([]string, 0, len(tdoc.Tokens))
	for _, tok := range tdoc.Tokens {
		lower := strings.ToLower(tok)

		if stopwordsES[lower] {
			continue // elimina stopword
		}
		lemma := applySuffixRules(lower)
		if lemma != "" {
			lemmas = append(lemmas, lemma)
		}
	}
	return LematizedDoc{TokenizedDoc: tdoc, Lemmas: lemmas}
}

// applySuffixRules reduce sufijos comunes del español.
// Retorna cadena vacía si el token es demasiado corto para ser útil.
func applySuffixRules(word string) string {
	if len(word) < 3 {
		return ""
	}
	suffixes := []struct{ suffix, replacement string }{
		{"aciones", "ar"},
		{"ación", "ar"},
		{"cion", "ar"},
		{"mente", ""},
		{"iendo", "er"},
		{"ando", "ar"},
		{"ados", "ar"},
		{"adas", "ar"},
		{"ado", "ar"},
		{"ada", "ar"},
		{"idos", "ir"},
		{"idas", "ir"},
		{"ido", "ir"},
		{"ida", "ir"},
		{"es", ""},
	}
	for _, s := range suffixes {
		if strings.HasSuffix(word, s.suffix) {
			base := strings.TrimSuffix(word, s.suffix) + s.replacement
			if len(base) >= 3 {
				return base
			}
		}
	}
	return word
}

// ── Pipeline secuencial principal ────────────────────────────────────────────

// RunSequential procesa el corpus de forma estrictamente secuencial.
// Por cada documento:
//  1. tokenize(doc)        → TokenizedDoc
//  2. lemmatize(tokenized) → LematizedDoc
//  3. Actualiza el estado global directamente.
//
// Retorna el estado final una vez procesado el corpus completo.
func RunSequential(corpus []Document) *PipelineState {
	state := &PipelineState{}

	for _, doc := range corpus {
		// ── Etapa 1: Tokenización BPE ────────────────────────
		tokenized := tokenize(doc)

		// ── Etapa 2: Lematización ────────────────────────────
		lematized := lemmatize(tokenized)

		state.TokensGlobales += len(tokenized.Tokens)
		state.DocsProcessados++
		if lematized.Origen == OrigenReal {
			state.DocsReales++
		} else {
			state.DocsSinteticos++
		}
	}

	return state
}

// ── Invariantes de corrección (equivale a los assert del Coordinador) ────────

// Verify comprueba los invariantes del pipeline sobre el estado final.
// Retorna una lista de violaciones (vacía si todo es correcto).
func Verify(state *PipelineState, nDocs, nReal, nSint int) []string {
	var violations []string

	if state.DocsProcessados != nDocs {
		violations = append(violations,
			"VIOLACIÓN: docs_procesados != N_DOCS (pérdida o duplicación)")
	}
	if state.DocsReales != nReal {
		violations = append(violations,
			"VIOLACIÓN: docs_reales != N_REAL (trazabilidad ORIGEN rota)")
	}
	if state.DocsSinteticos != nSint {
		violations = append(violations,
			"VIOLACIÓN: docs_sinteticos != N_SINT (trazabilidad ORIGEN rota)")
	}
	if state.DocsReales+state.DocsSinteticos != state.DocsProcessados {
		violations = append(violations,
			"VIOLACIÓN: docs_reales + docs_sinteticos != docs_procesados")
	}
	return violations
}
