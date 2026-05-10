// pipeline_test.go — Benchmarks del pipeline secuencial
//
// Ejecutar con:
//
//	go test -bench=. -benchmem -benchtime=10s
//	go test -bench=BenchmarkRunSequential -benchmem -run=^$
//	go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof
//	go tool pprof cpu.prof  (análisis de CPU)
//
// El flag -run=^$ evita ejecutar tests; solo corre benchmarks.
// El flag -benchmem muestra asignaciones de memoria (alocs/op).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const seqResultsDir = "../../resultados/seq_results/raw"
const finalDatasetPath = "../../data/dataset_final_1M.csv"

var benchmarkRunStamp = time.Now().Format("20060102-150405")
var benchmarkRunSeq uint64

// BenchmarkRunSequential mide el throughput del pipeline sobre corpus pequeños (100 docs).
// Utiliza corpus generado para aislar el rendimiento del pipeline de I/O.
func BenchmarkRunSequential(b *testing.B) {
	corpus := generateCorpus(50, 50) // 100 docs totales

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		RunSequential(corpus)
	}
	writeBenchmarkArtifact("BenchmarkRunSequential", "synthetic-100", 1, len(corpus), b.N, time.Since(start))
}

// BenchmarkRunSequential_Small mide el pipeline sobre 500 docs (pequeño).
func BenchmarkRunSequential_Small(b *testing.B) {
	corpus := generateCorpus(250, 250) // 500 docs

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		RunSequential(corpus)
	}
	writeBenchmarkArtifact("BenchmarkRunSequential_Small", "synthetic-500", 1, len(corpus), b.N, time.Since(start))
}

// BenchmarkRunSequential_Medium mide el pipeline sobre 10k docs (mediano).
func BenchmarkRunSequential_Medium(b *testing.B) {
	corpus := generateCorpus(5000, 5000) // 10k docs

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		RunSequential(corpus)
	}
	writeBenchmarkArtifact("BenchmarkRunSequential_Medium", "synthetic-10k", 1, len(corpus), b.N, time.Since(start))
}

// BenchmarkRunSequential_Large mide el pipeline sobre 100k docs (grande).
// Útil para análisis de escalabilidad.
func BenchmarkRunSequential_Large(b *testing.B) {
	corpus := generateCorpus(50000, 50000) // 100k docs

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		RunSequential(corpus)
	}
	writeBenchmarkArtifact("BenchmarkRunSequential_Large", "synthetic-100k", 1, len(corpus), b.N, time.Since(start))
}

// BenchmarkTokenize mide solo la etapa de tokenización BPE.
func BenchmarkTokenize(b *testing.B) {
	doc := Document{
		ID:      1,
		Sumilla: "Resolución que aprueba la transferencia de partidas del presupuesto institucional para gastos operacionales.",
		Origen:  OrigenReal,
	}

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		tokenize(doc)
	}
	writeBenchmarkArtifact("BenchmarkTokenize", "synthetic-stage", 1, b.N, b.N, time.Since(start))
}

// BenchmarkLemmatize mide solo la etapa de lematización.
func BenchmarkLemmatize(b *testing.B) {
	doc := Document{
		ID:      1,
		Sumilla: "Resolución que aprueba la transferencia de partidas del presupuesto institucional para gastos operacionales.",
		Origen:  OrigenReal,
	}
	tdoc := tokenize(doc)

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		lemmatize(tdoc)
	}
	writeBenchmarkArtifact("BenchmarkLemmatize", "synthetic-stage", 1, b.N, b.N, time.Since(start))
}

// BenchmarkApplySuffixRules mide el costo de la lematización heurística.
func BenchmarkApplySuffixRules(b *testing.B) {
	words := []string{
		"realización",
		"corriendo",
		"estudiando",
		"procesadas",
		"eliminación",
		"frecuentemente",
		"ingresos",
	}

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		for _, w := range words {
			applySuffixRules(w)
		}
	}
	writeBenchmarkArtifact("BenchmarkApplySuffixRules", "synthetic-stage", 1, len(words), b.N, time.Since(start))
}

// ── Benchmarks comparativos (para speedup análisis) ──────────────────────────

// BenchmarkRunSequential_FixedIter corre exactamente 1000 iteraciones
// (útil para comparar con versión concurrente con el mismo número de iteraciones).
func BenchmarkRunSequential_FixedIter(b *testing.B) {
	corpus := generateCorpus(122, 378) // ~500 docs (proporciones 244777:755223)

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		RunSequential(corpus)
	}
	writeBenchmarkArtifact("BenchmarkRunSequential_FixedIter", "synthetic-500", 1, len(corpus), b.N, time.Since(start))
}

// BenchmarkRunSequential_Final1M mide el pipeline sobre el dataset final de 1M registros.
// Este es el escenario comparable contra la versión concurrente.
func BenchmarkRunSequential_Final1M(b *testing.B) {
	corpus := loadBenchmarkCorpus(b, finalDatasetPath, defaultNReal, defaultNSint)

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		RunSequential(corpus)
	}
	runID := atomic.AddUint64(&benchmarkRunSeq, 1)
	writeBenchmarkArtifact("BenchmarkRunSequential_Final1M", "final-dataset-1m", 1, len(corpus), b.N, time.Since(start), runID)
}

// ── Helpers para benchmarks ──────────────────────────────────────────────────

// benchmarkResult agrupa métricas de un benchmark para análisis de speedup.
// Se usa en análisis posteriores: S(n) = T(1) / T(n)
type benchmarkResult struct {
	Timestamp    string  `json:"timestamp"`
	RunID        uint64  `json:"run_id,omitempty"`
	Dataset      string  `json:"dataset"`
	Name         string  `json:"name"`
	NWorkers     int     `json:"workers"`
	CorpusDocs   int     `json:"corpus_docs"`
	TotalOps     int     `json:"total_ops"`
	TotalDocs    int     `json:"total_docs"`
	ElapsedNanos int64   `json:"elapsed_nanos"`
	NsPerOp      float64 `json:"ns_per_op"`
	DocsPerSec   float64 `json:"docs_per_sec"`
	TimePerDoc   float64 `json:"ns_per_doc"`
}

func writeBenchmarkArtifact(name, dataset string, nWorkers, corpusDocs, totalOps int, elapsed time.Duration, runID ...uint64) {
	var currentRunID uint64
	if len(runID) > 0 {
		currentRunID = runID[0]
	}
	metrics := benchmarkResult{
		Timestamp:    benchmarkRunStamp,
		RunID:        currentRunID,
		Dataset:      dataset,
		Name:         name,
		NWorkers:     nWorkers,
		CorpusDocs:   corpusDocs,
		TotalOps:     totalOps,
		TotalDocs:    corpusDocs * totalOps,
		ElapsedNanos: elapsed.Nanoseconds(),
		NsPerOp:      float64(elapsed.Nanoseconds()) / float64(totalOps),
		DocsPerSec:   float64(corpusDocs*totalOps) / elapsed.Seconds(),
		TimePerDoc:   float64(elapsed.Nanoseconds()) / float64(corpusDocs*totalOps),
	}

	if err := persistBenchmarkResult(metrics); err != nil {
		fmt.Fprintf(os.Stderr, "[benchmark] no se pudo persistir %s: %v\n", name, err)
	}
	printBenchmarkReport(metrics)
}

func loadBenchmarkCorpus(b *testing.B, csvPath string, nReal, nSint int) []Document {
	b.Helper()
	b.StopTimer()
	corpus, err := LoadCorpus(csvPath, nReal, nSint)
	if err != nil {
		b.Fatalf("no se pudo cargar el corpus de benchmark: %v", err)
	}
	b.StartTimer()
	return corpus
}

// printBenchmarkReport imprime un resumen legible del benchmark.
func printBenchmarkReport(metrics benchmarkResult) {
	sep := strings.Repeat("─", 60)
	fmt.Printf("\n%s\n", sep)
	fmt.Println("  REPORTE DE BENCHMARK")
	fmt.Printf("%s\n", sep)
	fmt.Printf("  Test:           %s\n", metrics.Name)
	fmt.Printf("  Dataset:        %s\n", metrics.Dataset)
	if metrics.RunID > 0 {
		fmt.Printf("  Run ID:         %d\n", metrics.RunID)
	}
	fmt.Printf("  Workers:        %d\n", metrics.NWorkers)
	fmt.Printf("  Corpus docs:    %d\n", metrics.CorpusDocs)
	fmt.Printf("  Iteraciones:    %d\n", metrics.TotalOps)
	fmt.Printf("  Total docs:     %d\n", metrics.TotalDocs)
	fmt.Printf("  Tiempo total:   %.6f s\n", float64(metrics.ElapsedNanos)/1e9)
	fmt.Printf("  Ns/op:          %.2f\n", metrics.NsPerOp)
	fmt.Printf("  Throughput:     %.2f docs/s\n", metrics.DocsPerSec)
	fmt.Printf("  Tiempo/doc:     %.2f ns\n", metrics.TimePerDoc)
	fmt.Printf("%s\n\n", sep)
}

func persistBenchmarkResult(metrics benchmarkResult) error {
	if err := os.MkdirAll(seqResultsDir, 0o755); err != nil {
		return err
	}

	return writeBenchmarkJSON(metrics)
}

func writeBenchmarkJSON(metrics benchmarkResult) error {
	fileName := sanitizeBenchmarkName(metrics.Name)
	if metrics.RunID > 0 {
		fileName = fmt.Sprintf("%s_run_%02d", fileName, metrics.RunID)
	}
	fileName += ".json"
	path := filepath.Join(seqResultsDir, fileName)
	content, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(content, '\n'), 0o644)
}

func sanitizeBenchmarkName(name string) string {
	replacer := strings.NewReplacer(" ", "_", "-", "_", "(", "", ")", "", ":", "", "/", "_")
	return replacer.Replace(name)
}
