package main

import "github.com/josel/cc65_pc2/CC65_Diario_elPeruano_Grupo6/src/internal/nlp"

func nlpClean(s string) string       { return nlp.Clean(s) }
func nlpTokenize(s string) []string  { return nlp.Tokenize(s) }
func nlpLemmatize(tok string) string { return nlp.Lemmatize(tok) }
