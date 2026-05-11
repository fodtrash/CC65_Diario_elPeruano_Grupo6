package main

type PipelineState struct {
	DocsProcessados int
	DocsReales      int
	DocsSinteticos  int
	TokensGlobales  int
}

func tokenize(doc Document) TokenizedDoc {
	clean := nlpClean(doc.Sumilla)
	tokens := nlpTokenize(clean)
	return TokenizedDoc{Document: doc, Tokens: tokens}
}

func lemmatize(tdoc TokenizedDoc) LematizedDoc {
	lemmas := make([]string, len(tdoc.Tokens))
	for i, tok := range tdoc.Tokens {
		lemmas[i] = nlpLemmatize(tok)
	}
	return LematizedDoc{TokenizedDoc: tdoc, Lemmas: lemmas}
}

func RunSequential(corpus []Document) *PipelineState {
	state := &PipelineState{}
	for _, doc := range corpus {
		tokenized := tokenize(doc)
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

func Verify(state *PipelineState, nDocs, nReal, nSint int) []string {
	var violations []string
	if state.DocsProcessados != nDocs {
		violations = append(violations, "VIOLACIÓN: docs_procesados != N_DOCS")
	}
	if state.DocsReales != nReal {
		violations = append(violations, "VIOLACIÓN: docs_reales != N_REAL")
	}
	if state.DocsSinteticos != nSint {
		violations = append(violations, "VIOLACIÓN: docs_sinteticos != N_SINT")
	}
	if state.DocsReales+state.DocsSinteticos != state.DocsProcessados {
		violations = append(violations, "VIOLACIÓN: reales + sinteticos != procesados")
	}
	return violations
}
