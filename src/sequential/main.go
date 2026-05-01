// main.go — Punto de entrada del pipeline secuencial NLP
//
// Uso:
//
//	go run . [flags]
//	go run . -csv ../../data/sample/dataset_sample_500_rows.csv -n 500
//	go run . -nreal 244777 -nsint 755223          # corpus completo generado
//
// Flags:
//
//	-input ruta al CSV del corpus (default: ../../data/sample/dataset_sample_500_rows.csv)
//	-n     límite de documentos a leer del CSV (0 = sin límite)
//	-nreal documentos REAL  para corpus generado (ignorado si -csv existe)
//	-nsint documentos SINT  para corpus generado (ignorado si -csv existe)
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// ── Valores por defecto (fieles a las proporciones del proyecto) ─────────────

const (
	defaultCSV   = "../../data/sample/dataset_sample_500_rows.csv"
	defaultNReal = 244777
	defaultNSint = 755223
)

func main() {
	// ── Flags de configuración ───────────────────────────────────────────
	csvPath := flag.String("input", defaultCSV, "Ruta al CSV del corpus aumentado")
	nLimit := flag.Int("n", 0, "Límite de docs a leer del CSV (0 = todos)")
	nReal := flag.Int("nreal", defaultNReal, "Docs REAL para corpus generado")
	nSint := flag.Int("nsint", defaultNSint, "Docs SINT para corpus generado")
	flag.Parse()

	nDocs := *nReal + *nSint

	fmt.Println(banner())
	fmt.Printf("  Configuración  : nReal=%d | nSint=%d | nDocs=%d\n",
		*nReal, *nSint, nDocs)
	fmt.Printf("  Corpus CSV     : %s\n", *csvPath)
	if *nLimit > 0 {
		fmt.Printf("  Límite lectura : %d documentos\n", *nLimit)
	}
	fmt.Println(strings.Repeat("─", 60))

	// ── Carga del corpus ─────────────────────────────────────────────────
	corpus, err := LoadCorpus(*csvPath, *nReal, *nSint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}

	// Ajusta nReal/nSint si el corpus fue leído desde CSV
	actualNReal, actualNSint := countOrigens(corpus)
	if actualNReal+actualNSint == len(corpus) {
		*nReal = actualNReal
		*nSint = actualNSint
		nDocs = len(corpus)
	}

	// Aplica límite si fue especificado
	if *nLimit > 0 && *nLimit < len(corpus) {
		corpus = corpus[:*nLimit]
		nDocs = len(corpus)
		actualNReal, actualNSint = countOrigens(corpus)
		*nReal = actualNReal
		*nSint = actualNSint
	}

	fmt.Printf("[corpus] Corpus efectivo: %d docs (%d REAL + %d SINT)\n\n",
		nDocs, *nReal, *nSint)

	// ── Ejecución del pipeline secuencial ────────────────────────────────
	fmt.Println("  Iniciando pipeline secuencial…")
	start := time.Now()

	state := RunSequential(corpus)

	elapsed := time.Since(start)

	// ── Verificación de invariantes (assert del Coordinador Promela) ─────
	violations := Verify(state, nDocs, *nReal, *nSint)

	// ── Reporte de resultados ─────────────────────────────────────────────
	printReport(state, elapsed, violations, nDocs, *nReal, *nSint)

	if len(violations) > 0 {
		os.Exit(2)
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// countOrigens cuenta cuántos documentos tienen cada valor de ORIGEN.
func countOrigens(docs []Document) (nReal, nSint int) {
	for _, d := range docs {
		if d.Origen == OrigenReal {
			nReal++
		} else {
			nSint++
		}
	}
	return
}

// printReport imprime el reporte de ejecución con métricas y verificación.
func printReport(
	state *PipelineState,
	elapsed time.Duration,
	violations []string,
	nDocs, nReal, nSint int,
) {
	sep := strings.Repeat("─", 60)
	fmt.Println(sep)
	fmt.Println("  REPORTE DE EJECUCIÓN — Pipeline Secuencial")
	fmt.Println(sep)

	// Métricas de rendimiento
	docsPerSec := 0.0
	if elapsed.Seconds() > 0 {
		docsPerSec = float64(state.DocsProcessados) / elapsed.Seconds()
	}
	fmt.Printf("  %-30s %v\n", "Tiempo de ejecución:", elapsed)
	fmt.Printf("  %-30s %.2f docs/s\n", "Throughput:", docsPerSec)
	fmt.Printf("  %-30s %s\n", "Workers (pool Tok.):", "1 (secuencial)")
	fmt.Printf("  %-30s %s\n", "Workers (pool Lem.):", "1 (secuencial)")
	fmt.Println(sep)

	// Estado global (equivale a los contadores del modelo Promela)
	fmt.Printf("  %-30s %d / %d\n",
		"Docs procesados (esperado):", state.DocsProcessados, nDocs)
	fmt.Printf("  %-30s %d / %d\n",
		"Docs REAL (esperado):", state.DocsReales, nReal)
	fmt.Printf("  %-30s %d / %d\n",
		"Docs SINTETICO (esperado):", state.DocsSinteticos, nSint)
	fmt.Printf("  %-30s %d\n",
		"Tokens generados:", state.TokensGlobales)
	fmt.Println(sep)

	// Invariantes de corrección
	fmt.Println("  VERIFICACIÓN DE INVARIANTES")
	if len(violations) == 0 {
		fmt.Println("  ✓ Todos los invariantes se cumplen")
		fmt.Println("  ✓ Sin pérdida ni duplicación de documentos")
		fmt.Println("  ✓ Trazabilidad ORIGEN conservada en todo el pipeline")
	} else {
		for _, v := range violations {
			fmt.Printf("  ✗ %s\n", v)
		}
	}
	fmt.Println(sep)
}

// banner retorna el encabezado del programa.
func banner() string {
	return `
  ╔══════════════════════════════════════════════════════════╗
  ║  Pipeline NLP Secuencial — Diario Oficial El Peruano     ║
  ║  Etapas: Tokenización BPE → Lematización + Stopwords     ║
  ║  Baseline para análisis de Speedup y Scalability         ║
  ╚══════════════════════════════════════════════════════════╝`
}
