package nlp

import "strings"

type vocabEntry struct {
	canonical string
	vec       map[string]float64
}

var vocab []vocabEntry

var canonicalForms = []string{
	// Verbos de acción administrativa
	"designar", "nombrar", "aprobar", "autorizar", "disponer", "establecer",
	"modificar", "derogar", "suspender", "ampliar", "reducir", "transferir",
	"delegar", "reasignar", "encargar", "ratificar", "rectificar", "precisar",
	"declarar", "reconocer", "otorgar", "conceder", "denegar", "revocar",
	"aceptar", "rechazar", "admitir", "resolver", "ordenar", "regular",
	"incorporar", "excluir", "incluir", "crear", "disolver", "fusionar",
	"adscribir", "suprimir", "implementar", "ejecutar", "contratar", "renovar",
	"prorrogar", "convocar", "publicar", "notificar", "comunicar", "informar",
	// Sustantivos de documentos
	"resolucion", "decreto", "directiva", "reglamento", "norma", "ley",
	"ordenanza", "acuerdo", "convenio", "contrato", "oficio", "memorando",
	"informe", "dictamen", "opinion", "proyecto", "programa", "plan",
	"presupuesto", "expediente", "solicitud", "peticion", "recurso",
	"apelacion", "queja", "denuncia", "demanda", "sentencia",
	// Cargos y roles
	"director", "directora", "gerente", "jefe", "presidente", "ministro",
	"secretario", "coordinador", "supervisor", "inspector", "auditor",
	"asesor", "consultor", "especialista", "analista", "tecnico", "asistente",
	"subgerente", "subdirector", "viceministro", "titular", "funcionario",
	"servidor", "empleado", "trabajador", "contratado", "comisionado",
	// Entidades
	"ministerio", "oficina", "unidad", "area", "departamento", "division",
	"gerencia", "subgerencia", "direccion", "presidencia", "congreso",
	"municipalidad", "gobierno", "estado", "entidad", "organismo", "institucion",
	"empresa", "sociedad", "asociacion", "fundacion", "comision", "comite",
	"consejo", "tribunal", "juzgado", "fiscalia", "contraloria",
	// Términos jurídicos
	"vigencia", "plazo", "fecha", "periodo", "ejercicio", "gestion",
	"cargo", "puesto", "funcion", "competencia", "atribucion", "facultad",
	"obligacion", "responsabilidad", "derecho", "deber", "potestad",
	"jurisdiccion", "procedimiento", "tramite", "proceso", "instancia",
	"impugnacion", "revision", "control", "supervision", "fiscalizacion",
	"auditoria", "evaluacion", "seguimiento", "monitoreo",
	// Términos presupuestales
	"asignacion", "partida", "credito", "habilitacion", "anulacion",
	"devengado", "girado", "pagado", "saldo", "marco", "techo",
	"meta", "indicador", "resultado",
	// Adjetivos comunes
	"nacional", "regional", "local", "publico", "privado", "mixto",
	"temporal", "permanente", "provisional", "definitivo", "urgente",
	"prioritario", "especial", "general", "especifico", "administrativo",
	"ejecutivo", "legislativo", "judicial", "institucional", "sectorial",
}

func ngramVector(tok string) map[string]float64 {
	padded := "<" + tok + ">"
	runes := []rune(padded)
	counts := make(map[string]float64, len(runes)*2)

	// Bigramas de caracteres
	for i := 0; i < len(runes)-1; i++ {
		ng := string(runes[i : i+2])
		counts[ng]++
	}
	// Trigramas de caracteres
	for i := 0; i < len(runes)-2; i++ {
		ng := string(runes[i : i+3])
		counts[ng]++
	}

	// Normalización L2 con Newton-Raphson (evita import de math)
	var norm float64
	for _, v := range counts {
		norm += v * v
	}
	if norm > 0 {
		z := norm / 2
		for i := 0; i < 3; i++ {
			z -= (z*z - norm) / (2 * z)
		}
		invNorm := 1.0 / z
		for k := range counts {
			counts[k] *= invNorm
		}
	}
	return counts
}

func cosineSim(a, b map[string]float64) float64 {
	if len(b) < len(a) {
		a, b = b, a
	}
	var dot float64
	for k, va := range a {
		if vb, ok := b[k]; ok {
			dot += va * vb
		}
	}
	return dot
}

func suffixFallback(tok string) string {
	type rule struct {
		suffix  string
		minStem int
	}
	rules := []rule{
		{"amiento", 3}, {"imiento", 3}, {"amente", 3}, {"mente", 3},
		{"acion", 3}, {"icion", 3}, {"cion", 3}, {"sion", 3},
		{"ando", 3}, {"iendo", 3}, {"aron", 3}, {"ieron", 3},
		{"izar", 3}, {"idad", 3}, {"ado", 3}, {"ada", 3},
		{"ido", 3}, {"ida", 3}, {"oso", 3}, {"osa", 3},
	}
	for _, r := range rules {
		if strings.HasSuffix(tok, r.suffix) {
			stem := tok[:len(tok)-len(r.suffix)]
			if len(stem) >= r.minStem {
				return stem
			}
		}
	}
	return tok
}

func init() {
	vocab = make([]vocabEntry, len(canonicalForms))
	for i, form := range canonicalForms {
		vocab[i] = vocabEntry{
			canonical: form,
			vec:       ngramVector(form),
		}
	}
}

func Lemmatize(tok string) string {
	if len(tok) < 3 {
		return tok
	}

	tokVec := ngramVector(tok)

	const threshold = 0.42
	bestSim := 0.0
	bestForm := ""

	for i := range vocab {
		sim := cosineSim(tokVec, vocab[i].vec)
		if sim > bestSim {
			bestSim = sim
			bestForm = vocab[i].canonical
		}
	}

	if bestSim >= threshold {
		return bestForm
	}
	return suffixFallback(tok)
}
