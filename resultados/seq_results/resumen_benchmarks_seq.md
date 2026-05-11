# Tabla Maestra de Estadísticas — Pipeline Secuencial

Esta tabla resume el escenario comparable contra el pipeline concurrente: procesamiento end-to-end del dataset final de 1,000,000 registros. Las mediciones provienen de cinco ejecuciones controladas del benchmark `BenchmarkRunSequential_Final1M`, ejecutado como `go test -bench="BenchmarkRunSequential_Final1M" -benchmem -run="^$" -benchtime=1x -count=5`.

## Estadísticas por configuración

| Configuración | Mediciones (s) | Media | Media recortada | StdDev | CV% | Throughput recortado (docs/s) |
|:---|:---|---:|---:|---:|---:|---:|
| Final 1M, 1 worker | 586.987, 538.007, 588.707, 685.182, 665.086 | 612.794 | **613.593** | 60.856 | 9.9% | 1,629.74 |

Media recortada: se eliminan el valor mínimo (run 2: 538.007 s) y el máximo (run 4: 685.182 s) de las cinco mediciones y se promedian los tres valores centrales (runs 1, 3 y 5).

## Resultados individuales

| Run | Archivo JSON | Tiempo total (s) | Throughput (docs/s) | ns/doc |
|:---:|:---|---:|---:|---:|
| 1 | [BenchmarkRunSequential_Final1M_run_01.json](raw/BenchmarkRunSequential_Final1M_run_01.json) | 586.987 | 1,703.62 | 586,986.64 |
| 2 | [BenchmarkRunSequential_Final1M_run_02.json](raw/BenchmarkRunSequential_Final1M_run_02.json) | 538.007 | 1,858.71 | 538,006.73 |
| 3 | [BenchmarkRunSequential_Final1M_run_03.json](raw/BenchmarkRunSequential_Final1M_run_03.json) | 588.707 | 1,698.64 | 588,706.85 |
| 4 | [BenchmarkRunSequential_Final1M_run_04.json](raw/BenchmarkRunSequential_Final1M_run_04.json) | 685.182 | 1,459.47 | 685,181.62 |
| 5 | [BenchmarkRunSequential_Final1M_run_05.json](raw/BenchmarkRunSequential_Final1M_run_05.json) | 665.086 | 1,503.56 | 665,086.46 |

## Comparación preliminar con la referencia concurrente

Usando la media recortada secuencial de **613.593 s** como línea base, la comparación con la tabla concurrente proporcionada por el colega (cuyos tiempos están en **milisegundos**) puede expresarse como $T_{seq} / T_{conc}$.

> ⚠️ **Nota de unidades:** los benchmarks secuenciales reportan tiempos en segundos; los concurrentes en milisegundos. La conversión es necesaria para la comparación.

| Configuración concurrente | T_conc (s) | T_seq / T_conc | Speedup_conc_vs_seq |
|:---|---:|---:|---:|
| N=1, b=1000 (concurrente) | 9.007 | 68.12x | — |
| N=8, b=5000 (mejor config) | 4.670 | 131.39x | — |

Bajo esta convención, el pipeline secuencial tarda aproximadamente **68x más** que el concurrente con 1 worker y **131x más** que la mejor configuración concurrente (N=8, batch=5000). Esto invierte completamente la conclusión preliminar de documentos anteriores basados en mediciones previas: el pipeline secuencial actual es sustancialmente más lento porque **carga y procesa el dataset real desde el CSV** (`dataset_final_1M.csv`) en lugar de usar un corpus generado en memoria.

El speedup real de la versión concurrente respecto a la secuencial, tomando la mejor configuración concurrente (4.670 s) y la media recortada secuencial (613.593 s), es:

$$S = T_{seq} / T_{conc} = 613.593 \text{ s} / 4.670 \text{ s} \approx \mathbf{131.4x}$$

## Observación metodológica

La diferencia de escala entre los benchmarks anterior y actual (de ~3 s a ~613 s) se explica por un cambio sustancial en el código: la versión actual de `BenchmarkRunSequential_Final1M` llama a `loadBenchmarkCorpus` con `finalDatasetPath = "../../data/dataset_final_1M.csv"`, lo que implica lectura de disco, parseo de CSV y construcción del corpus antes de ejecutar el pipeline. El tiempo medido por `elapsed` cubre únicamente la ejecución de `RunSequential(corpus)` (el `b.ResetTimer()` se llama antes del loop de benchmark), pero el corpus de 1M registros reales introduce una carga de trabajo por documento significativamente mayor que la del corpus generado sintéticamente, donde las sumillas son frases cortas y repetitivas.

Los benchmarks de tokenización, lematización y reglas de sufijo se conservan como métricas auxiliares de diagnóstico pero no deben usarse como base principal para comparar speedup con la versión concurrente. La tabla principal de este documento está centrada en la misma carga de 1M registros para que la comparación sea lo más iso-funcional posible.
