# Resultados del Benchmark — PC2 Pipeline Concurrente

Esta carpeta contiene la evidencia experimental del trabajo concurrente del PC2.

## Estructura

```
resultados/
├── README.md                          # Este archivo
├── raw/                               # 35 JSONs con datos crudos de cada corrida
│   ├── n1_b1000_run{1..5}.json
│   ├── n2_b1000_run{1..5}.json
│   ├── n4_b1000_run{1..5}.json
│   ├── n8_b1000_run{1..5}.json
│   ├── n16_b1000_run{1..5}.json
│   ├── n8_b100_run{1..5}.json
│   └── n8_b5000_run{1..5}.json
└── analisis/
    ├── tabla_maestra.md               # Estadisticas consolidadas
    ├── parrafos_informe.md            # Parrafos pre-redactados para el informe
    └── figs/                          # Graficas PNG (se agregan manualmente)
```

## Como regenerar los datos raw

Si necesitas regenerar los 35 JSONs de `raw/` desde cero:

1. Asegurate de tener `data/dataset_final_1M.csv` (ver `data/README.md`).
2. Desde la raiz del repo, ejecuta:

```powershell
.\run_benchmarks.ps1
```

Genera los 35 JSONs en `resultados/raw/`.

## Convencion de nombres

`raw/n{W}_b{B}_run{R}.json`

- **W** = numero de workers por etapa (1, 2, 4, 8, 16)
- **B** = batch size en documentos (100, 1000, 5000)
- **R** = numero de repeticion (1 a 5)

Ejemplo: `n8_b1000_run3.json` = 8 workers, batch=1000, tercera repeticion.

## Configuraciones probadas

| Configuracion | Variable | Repeticiones |
|---|---|---|
| N = {1, 2, 4, 8, 16}, batch=1000 | Scalability vs N workers | 5 |
| N=8, batch = {100, 1000, 5000} | Efecto del batch size | 5 |

## Metricas registradas en cada JSON

Cada archivo JSON contiene 21 campos. Los mas relevantes para el analisis:

| Campo | Descripcion |
|---|---|
| `elapsed_total_ms` | Tiempo total de ejecucion del pipeline (metrica principal) |
| `elapsed_read_ms` | Tiempo en lectura del CSV |
| `elapsed_token_ms` | Tiempo de la etapa de tokenizacion (incluye limpieza) |
| `elapsed_lemma_ms` | Tiempo de la etapa de lematizacion |
| `peak_memory_mb` | Memoria pico durante la ejecucion |
| `mutex_contention_ms` | Tiempo acumulado esperando el mutex global |
| `tokens_globales` | Contador global protegido con `sync.Mutex` |
| `docs_procesados` | Total de documentos procesados (debe == 1,000,000) |
| `docs_reales` | Documentos reales del Diario El Peruano |
| `docs_sinteticos` | Documentos generados sinteticamente |

Los contadores `tokens_globales`, `docs_procesados`, `docs_reales` y `docs_sinteticos` son variables protegidas por `sync.Mutex` que mapean directamente a las variables del modelo Promela del PC1 y se usan para verificar las aserciones del `proctype Coordinador`.
