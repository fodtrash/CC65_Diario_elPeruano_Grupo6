# Codigo fuente

## Arquitectura

El pipeline usa dos patrones concurrentes combinados:

- **Pipeline:** 3 etapas conectadas por canales buffered (`chan []Document`)
- **Worker Pool:** cada etapa de procesamiento lanza N goroutines que consumen del mismo canal de entrada (fan-out/fan-in)

```
Lector (1 goroutine)
   │ chTokens
   ▼
WorkerTokenizacion (N goroutines)  ← incluye Clean() + Tokenize()
   │ chLemmas
   ▼
WorkerLematizacion (N goroutines)
   │ resultsCh
   ▼
Agregacion (main goroutine)
```

Cada canal transporta `[]Document` (lotes), no documentos sueltos. El tamano de lote es configurable via `-batch-size`.

La funcion `Clean()` (limpieza de texto) se ejecuta dentro de los workers de Tokenizacion como paso previo a `Tokenize()`, no como etapa separada del pipeline. Esto coincide con el modelo Promela del PC1, que no incluye un proctype de limpieza.

## Mapeo Promela a Go

El modelo Promela del PC1 tiene 6 proctypes. La implementacion Go omite `WorkerSintetizador` y `Mezclador` porque la sintesis de datos ya se hizo en Python.

| Promela proctype | Go (main.go) | Descripcion |
|---|---|---|
| `Lector` | Goroutine que lee CSV y crea batches | Lee y envia lotes a chTokens |
| `WorkerSintetizador` | Omitido (sintesis hecha en Python) | - |
| `Mezclador` | Omitido (un solo stream de datos) | - |
| `WorkerTokenizacion` | Pool de N goroutines + WaitGroup | Clean + Tokenize, actualiza tokens_globales |
| `WorkerLematizacion` | Pool de N goroutines + WaitGroup | Lemmatize, actualiza docs_procesados/reales/sinteticos |
| `Coordinador` | Main goroutine: agregacion + assertions | Verifica invariantes del Promela |

### Sincronizacion

| Promela | Go |
|---|---|
| `chan mutex = [1] of {bit}` | `sync.Mutex` (GlobalCounters.mu) |
| `mutex?_; tokens_globales++; mutex!1` | `gc.mu.Lock(); gc.tokensGlobales += n; gc.mu.Unlock()` |
| `chan wg_tok, wg_lem` | `sync.WaitGroup` por pool |
| `assert(docs_procesados == N_DOCS)` | `if gc.docsProcesados != totalDocsLeidos { log.Fatalf(...) }` |

## Flags CLI

| Flag | Default | Descripcion |
|---|---|---|
| `-input` | `data/dataset_final_1M.csv` | Ruta al CSV de entrada |
| `-workers-token` | `4` | Workers de tokenizacion (incluye limpieza) |
| `-workers-lemma` | `4` | Workers de lematizacion |
| `-batch-size` | `1000` | Documentos por lote |
| `-buffer` | `8` | Tamano del buffer de canales |
| `-output` | `resultados/concurrent_metrics.json` | Ruta del JSON de metricas |

## Estructura del JSON de metricas

```json
{
  "version": "concurrent",
  "input_file": "data/dataset_final_1M.csv",
  "workers_token": 4,
  "workers_lemma": 4,
  "batch_size": 1000,
  "total_docs": 1000000,
  "total_tokens": 8847793,
  "total_lemmas_unique": 45321,
  "elapsed_total_ms": 3123,
  "elapsed_read_ms": 2800,
  "elapsed_token_ms": 2900,
  "elapsed_lemma_ms": 2850,
  "peak_memory_mb": 37.2,
  "num_cpus": 8,
  "tokens_globales": 8847793,
  "docs_procesados": 1000000,
  "docs_reales": 244779,
  "docs_sinteticos": 755221,
  "mutex_contention_ms": 0.53
}
```

Los campos `tokens_globales`, `docs_procesados`, `docs_reales`, `docs_sinteticos` son contadores protegidos por `sync.Mutex` que espejean las variables globales del modelo Promela. `mutex_contention_ms` mide el tiempo total que las goroutines esperaron para adquirir el lock.

## Paquete internal/nlp/

Funciones NLP puras (sin goroutines, sin mutex, sin estado global). Disenadas para ser compartidas entre la version concurrente y la secuencial:

- `Clean(s string) string` - Lowercase, remocion de acentos, normalizacion
- `Tokenize(s string) []string` - Split + filtro de stopwords + min length
- `Lemmatize(tok string) string` - Reglas de sufijos del espanol (20 reglas)
