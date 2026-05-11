# Resultados del Benchmark — Pipeline Concurrente

Esta carpeta contiene la evidencia experimental del pipeline concurrente del PC2,
generada con el código actual de la rama `feature/concurrent` (commit `a94e206`,
que incorpora memoización local por worker en la etapa de lematización).

## Estructura

```
con_results/
├── README.md                     # Este archivo
├── raw/                          # 35 JSONs (7 configs × 5 reps)
│   ├── n1_b1000_run{1..5}.json
│   ├── n2_b1000_run{1..5}.json
│   ├── n4_b1000_run{1..5}.json
│   ├── n8_b1000_run{1..5}.json
│   ├── n16_b1000_run{1..5}.json
│   ├── n8_b100_run{1..5}.json
│   └── n8_b5000_run{1..5}.json
└── analisis/
    ├── tabla_maestra.md          # Estadísticas consolidadas y hallazgos
    └── parrafos_informe.md       # Párrafos pre-redactados para el informe
```

## Cómo regenerar los 35 JSONs

Los datos son reproducibles desde el binario actual. Desde la raíz del repo,
en PowerShell:

```powershell
# 1. Compilar el binario desde src/ (donde vive go.mod)
cd src
go build -o "$env:TEMP\conc_bench.exe" ./concurrent
cd ..

# 2. Iterar las 7 configuraciones × 5 repeticiones (~6 min sobre 1M docs)
$rawDir = "resultados/con_results/raw"
$configs = @(
    @{ name="n1_b1000";  token=1;  lemma=1;  batch=1000 },
    @{ name="n2_b1000";  token=2;  lemma=2;  batch=1000 },
    @{ name="n4_b1000";  token=4;  lemma=4;  batch=1000 },
    @{ name="n8_b1000";  token=8;  lemma=8;  batch=1000 },
    @{ name="n16_b1000"; token=16; lemma=16; batch=1000 },
    @{ name="n8_b100";   token=8;  lemma=8;  batch=100 },
    @{ name="n8_b5000";  token=8;  lemma=8;  batch=5000 }
)
foreach ($c in $configs) {
    for ($i = 1; $i -le 5; $i++) {
        & "$env:TEMP\conc_bench.exe" -input data/dataset_final_1M.csv `
            -workers-token $c.token -workers-lemma $c.lemma `
            -batch-size $c.batch -output "$rawDir/$($c.name)_run$i.json" `
            | Out-Null
        Write-Host "$($c.name) run $i ok"
    }
}
```

## Convención de nombres

`raw/n{W}_b{B}_run{R}.json`

- **W** = número de workers por etapa (1, 2, 4, 8, 16)
- **B** = batch size en documentos (100, 1000, 5000)
- **R** = número de repetición (1 a 5)

Ejemplo: `n4_b1000_run3.json` = 4 workers por etapa, batch=1000, tercera repetición.

## Configuraciones probadas

| Configuración | Variable estudiada | Repeticiones |
|---|---|---|
| N = {1, 2, 4, 8, 16}, batch=1000 | Escalabilidad en número de workers | 5 |
| N=8, batch = {100, 1000, 5000} | Efecto del tamaño de lote | 5 |

## Métricas registradas en cada JSON

Cada archivo contiene 18 campos. Los más relevantes para el análisis:

| Campo | Descripción |
|---|---|
| `elapsed_total_ms` | Tiempo total wall-clock del pipeline (métrica principal) |
| `elapsed_read_ms` | Wall-clock de la etapa de lectura del CSV |
| `elapsed_token_ms` | Wall-clock de la etapa de tokenización (incluye limpieza) |
| `elapsed_lemma_ms` | Wall-clock de la etapa de lematización |
| `peak_memory_mb` | Memoria pico durante la ejecución |
| `mutex_contention_ms` | Tiempo acumulado esperando el mutex global |
| `tokens_globales` | Contador global protegido con `sync.Mutex` |
| `docs_procesados` | Total de documentos procesados (debe == 1,000,000) |
| `docs_reales` | Documentos reales del Diario El Peruano (244,779) |
| `docs_sinteticos` | Documentos generados sintéticamente (755,221) |
| `total_lemmas_unique` | Lemas únicos encontrados (debe ser 38,201 en las 35 corridas) |

Los contadores `tokens_globales`, `docs_procesados`, `docs_reales` y `docs_sinteticos` son variables protegidas por `sync.Mutex` que mapean directamente a las variables del modelo Promela del PC1. Se utilizan para verificar las 3 aserciones del `proctype Coordinador` al final de cada corrida.

## Resumen de hallazgos

Detalle completo en [`analisis/tabla_maestra.md`](analisis/tabla_maestra.md). Resumen:

- **Configuración óptima**: N=4 workers, batch=1,000 → 5,515 ms, speedup 2.09×, 99 MB.
- **Speedup máximo vs secuencial**: 111.3× (T_seq = 613,593 ms ÷ T_N=4 = 5,515 ms).
- **El caché de lematización por worker (commit `a94e206`)** desplazó el cuello de botella a la lectura del CSV: agregar más de 4 workers degrada el rendimiento.
- **Contención de mutex despreciable** (<7 ms sobre >5,500 ms): valida el diseño de agregación local.
- **Determinismo verificado**: `total_tokens=8,847,793` y `total_lemmas_unique=38,201` idénticos en las 35 corridas.
