# Tabla Maestra de Estadísticas — Pipeline Secuencial

Esta tabla resume el escenario comparable contra el pipeline concurrente: procesamiento end-to-end del dataset final de 1,000,000 registros. Las mediciones provienen de seis ejecuciones del benchmark `BenchmarkRunSequential_Final1M`, ejecutadas con `-benchtime=6s -count=3`.

## Estadísticas por configuración

| Configuración | Mediciones (ns/doc) | Media | Media recortada | CV% | Throughput recortado (docs/s) | Peak Memory recortado (MB) |
|:---|:---|---:|---:|---:|---:|---:|
| Final 1M, 1 worker | 3010.10, 2971.79, 2973.26, 2989.91, 3024.94, 2992.80 | 2993.8 | **2991.52** | 0.9% | 334,280.02 | 460.24 |

Media recortada: se eliminan el valor mínimo y el máximo de las seis mediciones y se promedian los cuatro valores centrales.

## Resultados individuales

| Run | Archivo JSON | ns/doc | Throughput (docs/s) | Peak Memory (MB) | total_ops |
|:---:|:---|---:|---:|---:|---:|
| 1 | [BenchmarkRunSequential_Final1M_run_01.json](raw/BenchmarkRunSequential_Final1M_run_01.json) | 3,010.10 | 332,215.06 | 459.55 | 1 |
| 2 | [BenchmarkRunSequential_Final1M_run_02.json](raw/BenchmarkRunSequential_Final1M_run_02.json) | 2,971.79 | 336,497.40 | 460.05 | 2 |
| 3 | [BenchmarkRunSequential_Final1M_run_03.json](raw/BenchmarkRunSequential_Final1M_run_03.json) | 2,973.26 | 336,330.95 | 460.05 | 3 |
| 4 | [BenchmarkRunSequential_Final1M_run_04.json](raw/BenchmarkRunSequential_Final1M_run_04.json) | 2,989.91 | 334,458.15 | 460.30 | 1 |
| 5 | [BenchmarkRunSequential_Final1M_run_05.json](raw/BenchmarkRunSequential_Final1M_run_05.json) | 3,024.94 | 330,585.12 | 460.55 | 2 |
| 6 | [BenchmarkRunSequential_Final1M_run_06.json](raw/BenchmarkRunSequential_Final1M_run_06.json) | 2,992.80 | 334,135.38 | 460.55 | 1 |

## Comparación preliminar con la referencia concurrente

Usando la media recortada secuencial actualizada de 2,991.52 ns/doc (equivalente a ~2.992 s para 1M documentos) como línea base, la comparación con la tabla concurrente proporcionada por el colega puede expresarse como $T_{seq} / T_{conc}$. Bajo esa convención, el secuencial resulta menor que la configuración concurrente de 1 worker y batch=1000 (9.007 s), con una razón aproximada de 0.33x, y también menor que la mejor configuración concurrente reportada (4.670 s), con una razón aproximada de 0.64x. Las nuevas mediciones refuerzan la conclusión de que el pipeline secuencial se comporta como control experimental más rápido que la versión concurrente reportada, indicando que el overhead del paralelismo no es compensado por beneficios de throughput bajo estas condiciones de carga y hardware.

## Observación metodológica

Los benchmarks de tokenización, lematización y reglas de sufijo se conservan como métricas auxiliares de diagnóstico, pero no deben usarse como base principal para comparar speedup con la versión concurrente. La tabla principal de este documento está centrada en la misma carga de 1M registros para que la comparación sea lo más iso-funcional posible.

## Actualización reciente (runs 01-06 con -benchtime=6s -count=3)

Se ejecutaron seis ejecuciones del benchmark con parámetros `-benchtime=6s -count=3`, capturando cada corrida con su número de operaciones correspondiente. El parámetro `-count=3` ejecuta 3 iteraciones del benchmark, lo que explica por qué algunos runs registran `total_ops > 1`. La métrica normalizada `ns_per_doc` permanece consistente a través de todas las corridas, validando la estabilidad y reproducibilidad de la línea base secuencial. La media recortada de 2,991.52 ns/doc y el coeficiente de variación de 0.9% demuestran un comportamiento altamente estable del pipeline.
