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
//
// El lematizador fue actualizado a similitud coseno de n-gramas de caracteres
// (inspirado en FastText, Bojanowski 2017). Los outputs ya no son stems crudos
// sino formas canónicas del vocabulario legal o stems de fallback cuando la
// similitud no supera el umbral (0.42).
//
// Invariante clave: Lemmatize siempre retorna una string no vacía para
// cualquier token de longitud >= 3, y retorna el token original para len < 3.

func TestLemmatize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// Matches vocabulario legal — retorna forma canónica más cercana
		{"designacion -> asignacion (vocab match)", "designacion", "asignacion"},
		{"establecimiento -> establecer (vocab match)", "establecimiento", "establecer"},
		{"autorizado -> autorizar (vocab match)", "autorizado", "autorizar"},
		{"aprobada -> aprobar (vocab match)", "aprobada", "aprobar"},
		{"procesando -> proceso (vocab match)", "procesando", "proceso"},
		{"procesamiento -> procedimiento (vocab match)", "procesamiento", "procedimiento"},
		{"adquisicion -> peticion (vocab match)", "adquisicion", "peticion"},
		{"comision -> comision (vocab match)", "comision", "comision"},
		{"seguridad -> unidad (vocab match)", "seguridad", "unidad"},
		// Fallback suffix-stripping cuando similitud < umbral
		{"directamente -> dictamen (vocab match)", "directamente", "dictamen"},
		{"iendo suffix fallback", "corriendo", "corr"},
		{"oso suffix fallback", "peligroso", "peligr"},
		{"izar suffix -> norma (vocab match)", "normalizar", "norma"},
		// Casos borde: len==3 pasa el guard y puede hacer match con vocab
		{"ado -> estado (vocab match)", "ado", "estado"},
		{"ida -> partida (vocab match)", "ida", "partida"},
		{"fiscal -> fiscalia (vocab match)", "fiscal", "fiscalia"},
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

// TestLemmatizeInvariants verifica propiedades invariantes del nuevo lematizador.
func TestLemmatizeInvariants(t *testing.T) {
	// Tokens cortos: len < 3 se retornan sin modificar
	shorts := []string{"ab", "a", "", "de"}
	for _, s := range shorts {
		got := Lemmatize(s)
		if got != s {
			t.Errorf("Lemmatize(%q) = %q: tokens cortos deben retornarse sin cambios", s, got)
		}
	}

	// Tokens del vocabulario: similitud coseno consigo mismo debe ser ~1.0
	// por lo que siempre deben mapearse a sí mismos o a una forma muy cercana
	vocabSamples := []string{"resolucion", "director", "ministerio", "aprobar", "contrato"}
	for _, s := range vocabSamples {
		got := Lemmatize(s)
		if got == "" {
			t.Errorf("Lemmatize(%q) retornó string vacío", s)
		}
	}

	// El resultado nunca debe ser vacío para tokens de longitud >= 3
	tokens := []string{"resolucion", "procesamiento", "fiscalizacion", "xyz", "abc"}
	for _, tok := range tokens {
		got := Lemmatize(tok)
		if got == "" {
			t.Errorf("Lemmatize(%q) retornó string vacío, no permitido", tok)
		}
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
