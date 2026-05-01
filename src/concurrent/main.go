package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/josel/cc65_pc2/CC65_Diario_elPeruano_Grupo6/src/internal/nlp"
)

// Document representa un registro del dataset del Diario El Peruano.
type Document struct {
	FechaPublicacion string
	Entidad          string
	Dispositivo      string
	Numero           string
	Sumilla          string
	FechaCorte       string
	Origen           string
	SumillaClean     string
	Tokens           []string
	Lemmas           []string
}

// Metrics contiene las métricas de ejecución del pipeline.
type Metrics struct {
	Version         string  `json:"version"`
	InputFile       string  `json:"input_file"`
	WorkersToken    int     `json:"workers_token"`
	WorkersLemma    int     `json:"workers_lemma"`
	BatchSize       int     `json:"batch_size"`
	TotalDocs       int64   `json:"total_docs"`
	TotalTokens     int64   `json:"total_tokens"`
	TotalLemmasUniq int     `json:"total_lemmas_unique"`
	ElapsedTotalMs  int64   `json:"elapsed_total_ms"`
	ElapsedReadMs   int64   `json:"elapsed_read_ms"`
	ElapsedTokenMs  int64   `json:"elapsed_token_ms"`
	ElapsedLemmaMs  int64   `json:"elapsed_lemma_ms"`
	PeakMemoryMB     float64 `json:"peak_memory_mb"`
	NumCPUs          int     `json:"num_cpus"`
	TokensGlobales   int64   `json:"tokens_globales"`
	DocsProcesados   int64   `json:"docs_procesados"`
	DocsReales       int64   `json:"docs_reales"`
	DocsSinteticos   int64   `json:"docs_sinteticos"`
	MutexContentionMs float64 `json:"mutex_contention_ms"`
}

// localResult almacena los conteos locales de cada worker de lematización.
type localResult struct {
	tokenCount  int64
	lemmaCounts map[string]int
}

// GlobalCounters espejea las variables globales del modelo Promela:
//   Promela: int docs_procesados, docs_reales, docs_sinteticos, tokens_globales
//   Promela: chan mutex = [1] of { bit }  →  Go: sync.Mutex
// Los workers actualizan estos contadores bajo mutex, tal como el Promela
// usa mutex?_ / mutex!1 para proteger las actualizaciones.
type GlobalCounters struct {
	mu             sync.Mutex
	tokensGlobales int64 // Promela: tokens_globales
	docsProcesados int64 // Promela: docs_procesados
	docsReales     int64 // Promela: docs_reales
	docsSinteticos int64 // Promela: docs_sinteticos
	contentionNs   int64 // acumulador de tiempo esperando el mutex (nanosegundos)
}

func main() {
	// ── CLI flags ──
	inputFile := flag.String("input", "data/dataset_final_1M.csv", "Ruta al CSV de entrada")
	workersToken := flag.Int("workers-token", 4, "Número de workers de tokenización")
	workersLemma := flag.Int("workers-lemma", 4, "Número de workers de lematización")
	batchSize := flag.Int("batch-size", 1000, "Documentos por lote")
	bufferSize := flag.Int("buffer", 8, "Tamaño del buffer de canales")
	outputFile := flag.String("output", "resultados/concurrent_metrics.json", "Ruta del JSON de métricas")
	flag.Parse()

	log.Printf("Pipeline concurrente: input=%s workers(T=%d,L=%d) batch=%d buffer=%d",
		*inputFile, *workersToken, *workersLemma, *batchSize, *bufferSize)

	// ── Canales del pipeline (3 etapas: Lector → Tokenización → Lematización) ──
	chTokens := make(chan []Document, *bufferSize) // Lector → Tokenización
	chLemmas := make(chan []Document, *bufferSize) // Tokenización → Lematización

	// ── Timestamps atómicos para medir cada etapa ──
	var readStart, readEnd int64
	var tokenStart, tokenEnd int64
	var lemmaStart, lemmaEnd int64

	// Contador global de documentos (seteado por Lector)
	var docCount int64

	totalStart := time.Now()

	// Contadores globales protegidos con mutex (espejean modelo Promela)
	var gc GlobalCounters

	// ════════════════════════════════════════════════
	// ETAPA 1: Lector (1 goroutine — proctype Lector)
	// Lee el CSV y envía lotes directamente al pool de Tokenización.
	// ════════════════════════════════════════════════
	go func() {
		atomic.StoreInt64(&readStart, time.Now().UnixNano())
		defer func() {
			atomic.StoreInt64(&readEnd, time.Now().UnixNano())
			close(chTokens)
		}()

		f, err := os.Open(*inputFile)
		if err != nil {
			log.Fatalf("Error al abrir archivo: %v", err)
		}
		defer f.Close()

		reader := csv.NewReader(f)
		reader.LazyQuotes = true
		reader.FieldsPerRecord = 7

		// Saltar header
		if _, err := reader.Read(); err != nil {
			log.Fatalf("Error al leer header: %v", err)
		}

		batch := make([]Document, 0, *batchSize)
		var total int64

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("WARN: fila malformada (doc %d): %v", total+1, err)
				continue
			}

			batch = append(batch, Document{
				FechaPublicacion: record[0],
				Entidad:          record[1],
				Dispositivo:      record[2],
				Numero:           record[3],
				Sumilla:          record[4],
				FechaCorte:       record[5],
				Origen:           record[6],
			})
			total++

			if len(batch) == *batchSize {
				chTokens <- batch
				batch = make([]Document, 0, *batchSize)
			}
		}
		// Enviar lote parcial restante
		if len(batch) > 0 {
			chTokens <- batch
		}
		atomic.StoreInt64(&docCount, total)
		log.Printf("Lector: %d documentos leídos", total)
	}()

	// ══════════════════════════════════════════════════════════════════
	// ETAPA 2: WorkerTokenizacion (N goroutines + WaitGroup)
	// Cada worker aplica Clean() + Tokenize() sobre la Sumilla cruda.
	// La limpieza (normalización de texto) se integra aquí como paso
	// previo a la tokenización, no como etapa separada del pipeline.
	// ══════════════════════════════════════════════════════════════════
	var wgToken sync.WaitGroup
	wgToken.Add(*workersToken)
	for i := 0; i < *workersToken; i++ {
		go func() {
			defer wgToken.Done()
			for batch := range chTokens {
				atomic.CompareAndSwapInt64(&tokenStart, 0, time.Now().UnixNano())
				var batchTokens int64
				for j := range batch {
					// Limpieza: lowercase, acentos, regex normalize
					batch[j].SumillaClean = nlp.Clean(batch[j].Sumilla)
					// Tokenización: split + stopwords + min length
					batch[j].Tokens = nlp.Tokenize(batch[j].SumillaClean)
					batchTokens += int64(len(batch[j].Tokens))
				}
				// Promela: mutex?_; tokens_globales++; mutex!1
				lockStart := time.Now()
				gc.mu.Lock()
				gc.tokensGlobales += batchTokens
				gc.mu.Unlock()
				atomic.AddInt64(&gc.contentionNs, time.Since(lockStart).Nanoseconds())
				chLemmas <- batch
			}
		}()
	}
	go func() {
		wgToken.Wait()
		atomic.StoreInt64(&tokenEnd, time.Now().UnixNano())
		close(chLemmas)
	}()

	// ══════════════════════════════════════════════════════════════
	// ETAPA 3: WorkerLematizacion (N goroutines + agregación local)
	// ══════════════════════════════════════════════════════════════
	resultsCh := make(chan localResult, *workersLemma)
	var wgLemma sync.WaitGroup
	wgLemma.Add(*workersLemma)
	for i := 0; i < *workersLemma; i++ {
		go func() {
			defer wgLemma.Done()
			local := localResult{lemmaCounts: make(map[string]int)}
			for batch := range chLemmas {
				atomic.CompareAndSwapInt64(&lemmaStart, 0, time.Now().UnixNano())
				// Contadores locales del batch para minimizar tiempo bajo mutex
				var batchDocs, batchReales, batchSinteticos int64
				for j := range batch {
					lemmas := make([]string, len(batch[j].Tokens))
					for k, tok := range batch[j].Tokens {
						lemma := nlp.Lemmatize(tok)
						lemmas[k] = lemma
						local.lemmaCounts[lemma]++
					}
					batch[j].Lemmas = lemmas
					local.tokenCount += int64(len(batch[j].Tokens))
					// Conteo por origen (Promela: if origen==REAL -> docs_reales++ :: else -> docs_sinteticos++)
					batchDocs++
					if batch[j].Origen == "REAL" {
						batchReales++
					} else {
						batchSinteticos++
					}
				}
				// Promela: mutex?_; docs_procesados++; docs_reales|sinteticos++; mutex!1
				lockStart := time.Now()
				gc.mu.Lock()
				gc.docsProcesados += batchDocs
				gc.docsReales += batchReales
				gc.docsSinteticos += batchSinteticos
				gc.mu.Unlock()
				atomic.AddInt64(&gc.contentionNs, time.Since(lockStart).Nanoseconds())
			}
			resultsCh <- local
		}()
	}
	go func() {
		wgLemma.Wait()
		atomic.StoreInt64(&lemmaEnd, time.Now().UnixNano())
		close(resultsCh)
	}()

	// ══════════════════════════════════════════════
	// AGREGACIÓN (goroutine principal — Coordinador)
	// ══════════════════════════════════════════════
	var totalTokens int64
	globalLemmas := make(map[string]int)
	for lr := range resultsCh {
		totalTokens += lr.tokenCount
		for lemma, count := range lr.lemmaCounts {
			globalLemmas[lemma] += count
		}
	}

	elapsed := time.Since(totalStart)

	// ══════════════════════════════════════════════════════════════════
	// ASERCIONES DEL COORDINADOR (Promela: proctype Coordinador)
	// Espejean las assert() del modelo formal validado en PC1.
	// ══════════════════════════════════════════════════════════════════
	totalDocsLeidos := atomic.LoadInt64(&docCount)

	// Promela: assert(docs_procesados == N_DOCS)
	if gc.docsProcesados != totalDocsLeidos {
		log.Fatalf("Assertion fallida [Promela: assert(docs_procesados==N_DOCS)]: docs_procesados=%d != N_DOCS=%d",
			gc.docsProcesados, totalDocsLeidos)
	}
	// Promela: assert(docs_reales + docs_sinteticos == docs_procesados)
	if gc.docsReales+gc.docsSinteticos != gc.docsProcesados {
		log.Fatalf("Assertion fallida [Promela: assert(docs_reales+docs_sinteticos==docs_procesados)]: reales(%d)+sinteticos(%d)=%d != procesados(%d)",
			gc.docsReales, gc.docsSinteticos, gc.docsReales+gc.docsSinteticos, gc.docsProcesados)
	}
	// Promela: consistencia tokens_globales con conteo local
	if gc.tokensGlobales != totalTokens {
		log.Fatalf("Assertion fallida [tokens_globales]: mutex_counter=%d != local_aggregation=%d",
			gc.tokensGlobales, totalTokens)
	}

	log.Printf("Assertions del Coordinador OK: docs_procesados=%d, reales=%d, sinteticos=%d, tokens_globales=%d",
		gc.docsProcesados, gc.docsReales, gc.docsSinteticos, gc.tokensGlobales)

	// ── Métricas de memoria ──
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// ── Construir métricas ──
	metrics := Metrics{
		Version:           "concurrent",
		InputFile:         *inputFile,
		WorkersToken:      *workersToken,
		WorkersLemma:      *workersLemma,
		BatchSize:         *batchSize,
		TotalDocs:         totalDocsLeidos,
		TotalTokens:       totalTokens,
		TotalLemmasUniq:   len(globalLemmas),
		ElapsedTotalMs:    elapsed.Milliseconds(),
		ElapsedReadMs:     (atomic.LoadInt64(&readEnd) - atomic.LoadInt64(&readStart)) / 1e6,
		ElapsedTokenMs:    (atomic.LoadInt64(&tokenEnd) - atomic.LoadInt64(&tokenStart)) / 1e6,
		ElapsedLemmaMs:    (atomic.LoadInt64(&lemmaEnd) - atomic.LoadInt64(&lemmaStart)) / 1e6,
		PeakMemoryMB:      float64(ms.Sys) / (1024 * 1024),
		NumCPUs:           runtime.NumCPU(),
		TokensGlobales:    gc.tokensGlobales,
		DocsProcesados:    gc.docsProcesados,
		DocsReales:        gc.docsReales,
		DocsSinteticos:    gc.docsSinteticos,
		MutexContentionMs: float64(atomic.LoadInt64(&gc.contentionNs)) / 1e6,
	}

	// ── Escribir JSON ──
	outJSON, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Fatalf("Error al serializar métricas: %v", err)
	}

	// Imprimir a stdout
	fmt.Println(string(outJSON))

	// Escribir a archivo
	if err := os.MkdirAll("resultados", 0755); err != nil {
		log.Printf("WARN: no se pudo crear directorio resultados: %v", err)
	}
	if err := os.WriteFile(*outputFile, outJSON, 0644); err != nil {
		log.Printf("WARN: no se pudo escribir %s: %v", *outputFile, err)
	} else {
		log.Printf("Métricas escritas en %s", *outputFile)
	}

	// ── Resumen final ──
	log.Printf("Resumen: docs=%d tokens=%d lemmas_únicos=%d tiempo=%dms",
		metrics.TotalDocs, metrics.TotalTokens, metrics.TotalLemmasUniq, metrics.ElapsedTotalMs)
}
