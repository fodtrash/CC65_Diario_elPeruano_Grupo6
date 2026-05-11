# CC65 - PC2: Pipeline NLP del Diario Oficial El Peruano

Pipeline de preprocesamiento NLP para 1,000,000 de registros legales en espanol, con dos implementaciones en Go:

- `src/concurrent/`: pipeline concurrente con goroutines, channels, `sync.Mutex` y `sync.WaitGroup`
- `src/sequential/`: pipeline secuencial de referencia para comparacion, verificacion y baseline

El proyecto incluye limpieza, tokenizacion, lematizacion, benchmarks y salidas JSON para analisis de rendimiento.

## Integrantes

| Integrante | Codigo |
|---|---|
| Joaquin Sebastian Ruiz Ramirez | 20201F678 |
| Ricardo Martin Tejada Ramirez | 202113697 |
| Jose Giovanni Laura Silvera | 202112986 |

## Estructura del repositorio

```text
.
├── README.md
├── data/
│   ├── README.md
│   ├── dataset_final_1M.csv
│   └── sample/
│       └── dataset_sample_500_rows.csv
├── notebooks/
│   └── consolidacion_dataset_aumentado.ipynb
├── resultados/
│   ├── con_results/
│   │   ├── README.md
│   │   ├── analisis/
│   │   │   ├── parrafos_informe.md
│   │   │   ├── tabla_maestra.md
│   │   │   └── figs/
│   │   └── raw/
│   └── seq_results/
│       ├── analisis_benchmark_seq.md
│       ├── resumen_benchmarks_seq.md
│       └── raw/
└── src/
    ├── go.mod
    ├── concurrent/
    │   ├── main.go
    │   └── run_benchmarks.ps1
    ├── internal/
    │   └── nlp/
    │       ├── clean.go
    │       ├── tokenize.go
    │       ├── lemmatize.go
    │       └── nlp_test.go
    └── sequential/
        ├── corpus.go
        ├── main.go
        ├── pipeline.go
        └── pipeline_test.go
```

## Prerequisitos

- Go 1.22 o superior
- PowerShell 5.1 o superior para ejecutar los benchmarks en Windows
- Dataset generado en `data/dataset_final_1M.csv`

## Obtener el dataset

El dataset completo se genera desde `notebooks/consolidacion_dataset_aumentado.ipynb` y luego se guarda en `data/dataset_final_1M.csv`.
Para pruebas rapidas tambien existe `data/sample/dataset_sample_500_rows.csv`.

## Como ejecutar los pipelines

Todos los comandos Go se ejecutan desde `src/`, porque ahi vive el `go.mod`.

### Pipeline concurrente

```powershell
cd src
go run ./concurrent/ -input ../data/dataset_final_1M.csv
```

Con configuracion personalizada:

```powershell
cd src
go run ./concurrent/ `
	-input ../data/dataset_final_1M.csv `
	-workers-token 8 -workers-lemma 8 `
	-batch-size 1000 -buffer 8 `
	-output ../resultados/con_results/concurrent_metrics.json
```

Para dataset de prueba:

```powershell
cd src
go run ./concurrent/ -input ../data/sample/dataset_sample_500_rows.csv
```

### Pipeline secuencial

```powershell
cd src\sequential
go run . -input ../../data/dataset_final_1M.csv
```

Para limitar la lectura a N documentos:

```powershell
cd src\sequential
go run . -input ../../data/sample/dataset_sample_500_rows.csv -n 500
```

## Benchmarks

El runner automatiza 7 configuraciones con 5 repeticiones cada una y deja los JSON en `resultados/con_results/raw/`.

```powershell
cd src
.\concurrent\run_benchmarks.ps1
```

Configuraciones incluidas:

- `n1_b1000`
- `n2_b1000`
- `n4_b1000`
- `n8_b1000`
- `n16_b1000`
- `n8_b100`
- `n8_b5000`

## Tests

```powershell
cd src
go test ./...
```

## Verificacion con race detector

```powershell
cd src
go run -race ./concurrent/ -input ../data/sample/dataset_sample_500_rows.csv
```

## Resultados

Speedup medido sobre 1,000,000 documentos usando media recortada de 5 repeticiones por configuracion:

| Workers | Batch | Tiempo (ms) | Speedup |
|---|---|---|---|
| 1 | 1000 | 9007 | 1.00x |
| 2 | 1000 | 5506 | 1.64x |
| 4 | 1000 | 5504 | 1.64x |
| 8 | 1000 | 4878 | **1.85x** |
| 16 | 1000 | 4790 | 1.88x |
| 8 | 100 | 4934 | 1.83x |
| 8 | 5000 | 4670 | 1.93x |

- Speedup optimo: 1.85x con 8 workers y batch de 1000
- Mejor resultado observado: 1.93x con batch de 5000
- Hardware: 8 CPUs logicos, Windows 10/11
- Race detector: limpio, sin data races
- Determinismo: totales identicos en las 35 corridas, con `tokens=8,847,793` y `lemmas_unique=45,321`

Las salidas del pipeline concurrente se guardan en `resultados/con_results/`, incluyendo:

- `raw/`: JSON crudos de cada corrida
- `analisis/`: resumenes y tablas maestras

Los archivos pueden regenerarse desde los comandos anteriores.

## Verificacion formal

El modelo Promela y la justificacion bibliografica (Siino 2024, Treviso 2023, You 2025) estan documentados en el informe del PC1. La implementacion Go es una traduccion fiel de ese modelo:

- Cada `proctype` del modelo Promela mapea a un pool de goroutines en Go
- `chan mutex = [1] of {bit}` se implementa con `sync.Mutex`
- `chan wg_tok, wg_lem` se implementan con `sync.WaitGroup`
- Las tres aserciones del `proctype Coordinador` se verifican en tiempo de ejecucion

La verificacion formal con Spin se realizara en el TP de la semana 7.

## Notas de implementacion

- `src/internal/nlp/` contiene funciones NLP puras, reutilizadas por los pipelines
