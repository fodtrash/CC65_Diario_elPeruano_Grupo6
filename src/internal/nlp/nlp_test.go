package nlp

import (
	"reflect"
	"testing"
)

// --------------- Clean tests ---------------

func TestClean(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lowercase", "Designan Directora Ejecutiva", "designan directora ejecutiva"},
		{"accents", "Resolución Núm 068-2020", "resolucion num 068 2020"},
		{"special chars", "Aprobación de la ACCIÓN!!", "aprobacion de la accion"},
		{"multi spaces", "  multiple   spaces  ", "multiple spaces"},
		{"empty", "", ""},
		{"numbers preserved", "Artículo 123 del 2020", "articulo 123 del 2020"},
		{"ñ removal", "Diseño año España", "diseno ano espana"},
		{"ü removal", "pingüino bilingüe", "pinguino bilingue"},
		{"only symbols", "---!!!", ""},
		{"mixed", "  RESOLUCIÓN N° 0051/RE-2019  ", "resolucion n 0051 re 2019"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clean(tt.in)
			if got != tt.want {
				t.Errorf("Clean(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// --------------- Tokenize tests ---------------

func TestTokenize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			"normal",
			"designan directora ejecutiva del programa",
			[]string{"designan", "directora", "ejecutiva", "programa"},
		},
		{
			"all stopwords",
			"de la el los en por para con",
			[]string{},
		},
		{
			"short tokens filtered",
			"a b cd efg",
			[]string{"cd", "efg"},
		},
		{
			"empty",
			"",
			[]string{},
		},
		{
			"single valid token",
			"resolucion",
			[]string{"resolucion"},
		},
		{
			"mixed stopwords and valid",
			"nombran director de la oficina",
			[]string{"nombran", "director", "oficina"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tokenize(tt.in)
			if len(got) == 0 && len(tt.want) == 0 {
				return // both empty, ok
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tokenize(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// --------------- Lemmatize tests ---------------

func TestLemmatize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"acion suffix", "designacion", "design"},
		{"imiento suffix", "establecimiento", "establec"},
		{"amente suffix", "directamente", "direct"},
		{"ado suffix", "autorizado", "autoriz"},
		{"ada suffix", "aprobada", "aprob"},
		{"ando suffix", "procesando", "proces"},
		{"iendo suffix", "corriendo", "corr"},
		{"idad suffix", "seguridad", "segur"},
		{"no match", "fiscal", "fiscal"},
		{"short word no strip", "ado", "ado"},
		{"short word no strip 2", "ida", "ida"},
		{"amiento suffix", "procesamiento", "proces"},
		{"oso suffix", "peligroso", "peligr"},
		{"izar suffix", "normalizar", "normal"},
		{"icion suffix", "adquisicion", "adquis"},
		{"sion suffix", "comision", "comi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Lemmatize(tt.in)
			if got != tt.want {
				t.Errorf("Lemmatize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// --------------- Benchmarks ---------------

var sampleSumilla = "designan directora ejecutiva del programa nacional de apoyo directo a los mas pobres juntos"

func BenchmarkClean(b *testing.B) {
	raw := "Designan Directora Ejecutiva del Programa Nacional de Apoyo Directo a los Más Pobres JUNTOS"
	for i := 0; i < b.N; i++ {
		Clean(raw)
	}
}

func BenchmarkTokenize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Tokenize(sampleSumilla)
	}
}

func BenchmarkLemmatize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Lemmatize("procesamiento")
	}
}
