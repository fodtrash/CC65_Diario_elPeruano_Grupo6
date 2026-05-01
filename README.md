# CC65 - PC2: Pipeline Concurrente NLP del Diario Oficial El Peruano

Pipeline de preprocesamiento NLP (limpieza, tokenizacion, lematizacion) sobre 1,000,000 de registros legales en espanol, implementado con concurrencia en Go usando goroutines, channels, sync.Mutex y sync.WaitGroup.

**Curso:** CC65 - Programacion Concurrente y Distribuida
**Profesor:** Carlos Alberto Jara Garcia
**Semestre:** 2026-01

## Integrantes

| Integrante | Codigo |
|---|---|
| Joaquin Sebastian Ruiz Ramirez | 20201F678 |
| Ricardo Martin Tejada Ramirez | 202113697 |
| Jose Giovanni Laura Silvera | 202112986 |

## Estructura del repositorio

```
.
├── data/                          # Datasets
│   ├── dataset_final_1M.csv       # 1M registros (no en Git, ~200 MB)
│   ├── README.md                  # Como obtener el dataset
│   └── sample/
│       └── sample_500.csv         # 500 registros para tests rapidos
├── notebooks/
│   └── consolidacion_dataset_aumentado.ipynb  # Generacion del dataset
├── src/
│   ├── internal/nlp/              # Funciones NLP puras (sin concurrencia)
│   │   ├── clean.go               # Limpieza de texto
│   │   ├── tokenize.go            # Tokenizacion + stopwords
│   │   ├── lemmatize.go           # Lematizacion por sufijos
│   │   └── nlp_test.go            # 28 tests unitarios + benchmarks
│   ├── concurrent/
│   │   └── main.go                # Pipeline concurrente (3 etapas)
│   ├── sequential/                # (pendiente - Joaquin)
│   └── benchmarks/                # (pendiente - Ricardo)
├── resultados/                    # Metricas JSON (no en Git, regenerables)
│   └── raw/                       # 35 archivos: 7 configs x 5 repeticiones
├── run_benchmarks.ps1             # Script de 35 corridas automatizadas
├── go.mod                         # Modulo Go (solo stdlib)
└── .gitignore
```

## Como correr el codigo

### Prerequisitos

- Go 1.22 o superior
- Dataset generado (ver `data/README.md`)

### Obtener el dataset

Ejecutar el notebook `notebooks/consolidacion_dataset_aumentado.ipynb` en Google Colab y descargar `dataset_final_1M.csv` a la carpeta `data/`.

### Pipeline concurrente (dataset completo)

```bash
go run ./src/concurrent/ -input data/dataset_final_1M.csv
```

### Pipeline concurrente (sample rapido)

```bash
go run ./src/concurrent/ -input data/sample/sample_500.csv
```

### Configurar workers y batch size

```bash
go run ./src/concurrent/ \
  -input data/dataset_final_1M.csv \
  -workers-token 8 -workers-lemma 8 \
  -batch-size 1000 -buffer 8 \
  -output resultados/mi_test.json
```

### Correr benchmarks completos (7 configs x 5 repeticiones)

```powershell
.\run_benchmarks.ps1
```

### Correr tests unitarios

```bash
go test -v ./src/internal/nlp/
```

### Verificar con race detector

```bash
go run -race ./src/concurrent/ -input data/sample/sample_500.csv
```

## Resultados

Speedup medido sobre 1,000,000 documentos (media recortada, 5 repeticiones por configuracion):

| Workers | Batch | Tiempo (ms) | Speedup |
|---|---|---|---|
| 1 | 1000 | 9007 | 1.00x |
| 2 | 1000 | 5506 | 1.64x |
| 4 | 1000 | 5504 | 1.64x |
| 8 | 1000 | 4878 | **1.85x** |
| 16 | 1000 | 4790 | 1.88x |
| 8 | 100 | 4934 | 1.83x |
| 8 | 5000 | 4670 | 1.93x |

- **Speedup optimo:** 1.85x con 8 workers y batch de 1000 (1.93x con batch=5000)
- **Hardware:** 8 CPUs logicos, Windows 10/11
- **Race detector:** limpio (0 data races)
- **Determinismo:** totales identicos en las 35 corridas (tokens=8,847,793, lemmas_unique=45,321)

## Verificacion formal

El modelo Promela y la justificacion bibliografica (Siino 2024, Treviso 2023, You 2025) estan documentados en el informe del PC1. La implementacion Go es traduccion fiel de ese modelo:

- Cada `proctype` del Promela mapea a un pool de goroutines en Go
- `chan mutex = [1] of {bit}` mapea a `sync.Mutex`
- `chan wg_tok, wg_lem` mapean a `sync.WaitGroup`
- Las 3 aserciones del `proctype Coordinador` estan implementadas como checks en tiempo de ejecucion

La verificacion formal con Spin se realizara en el TP (semana 7).
