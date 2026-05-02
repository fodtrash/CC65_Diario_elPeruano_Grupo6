# Tabla Maestra de Estadísticas — Pipeline Secuencial

Esta tabla resume el escenario comparable contra el pipeline concurrente: procesamiento end-to-end del dataset final de 1,000,000 registros. Las mediciones provienen de cinco ejecuciones controladas del benchmark `BenchmarkRunSequential_Final1M`, ejecutado como `go test -bench="BenchmarkRunSequential_Final1M" -benchmem -run="^$" -benchtime=1x -count=5`.

## Estadísticas por configuración

| Configuración | Mediciones (s) | Media | Media recortada | StdDev | CV% | Throughput recortado (docs/s) |
|:---|:---|---:|---:|---:|---:|---:|
| Final 1M, 1 worker | 2.912, 3.382, 3.065, 2.917, 2.933 | 3.042 | **2.972** | 0.200 | 6.6% | 336,518.25 |

Media recortada: se eliminan el valor mínimo y el máximo de las cinco mediciones y se promedian los tres valores centrales.

## Resultados individuales

| Run | Archivo JSON | Tiempo total (s) | Throughput (docs/s) | ns/doc |
|:---:|:---|---:|---:|---:|
| 1 | [BenchmarkRunSequential_Final1M_run_01.json](raw/BenchmarkRunSequential_Final1M_run_01.json) | 2.912 | 343,350.78 | 2,912.47 |
| 2 | [BenchmarkRunSequential_Final1M_run_02.json](raw/BenchmarkRunSequential_Final1M_run_02.json) | 3.382 | 295,655.89 | 3,382.31 |
| 3 | [BenchmarkRunSequential_Final1M_run_03.json](raw/BenchmarkRunSequential_Final1M_run_03.json) | 3.065 | 326,289.96 | 3,064.76 |
| 4 | [BenchmarkRunSequential_Final1M_run_04.json](raw/BenchmarkRunSequential_Final1M_run_04.json) | 2.917 | 342,778.57 | 2,917.34 |
| 5 | [BenchmarkRunSequential_Final1M_run_05.json](raw/BenchmarkRunSequential_Final1M_run_05.json) | 2.933 | 340,979.48 | 2,932.73 |

## Comparación preliminar con la referencia concurrente

Usando la media recortada secuencial de 2.972 s como línea base, la comparación con la tabla concurrente proporcionada por el colega puede expresarse como $T_{seq} / T_{conc}$. Bajo esa convención, el secuencial resulta menor que la configuración concurrente de 1 worker y batch=1000 (9.007 s), con una razón aproximada de 0.33x, y también menor que la mejor configuración concurrente reportada (4.670 s), con una razón aproximada de 0.64x. Por tanto, en las mediciones actuales el pipeline secuencial se comporta como control experimental más rápido que la versión concurrente reportada, lo que obliga a interpretar el paralelismo como un costo adicional si no se igualan exactamente el trabajo por documento y el modelo de sincronización.

## Observación metodológica

Los benchmarks de tokenización, lematización y reglas de sufijo se conservan como métricas auxiliares de diagnóstico, pero no deben usarse como base principal para comparar speedup con la versión concurrente. La tabla principal de este documento está centrada en la misma carga de 1M registros para que la comparación sea lo más iso-funcional posible.
