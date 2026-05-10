# Análisis técnico-académico de los benchmarks secuenciales

La línea base secuencial fue reevaluada sobre el dataset final de 1,000,000 registros para asegurar una comparación más directa con la versión concurrente. Se ejecutó el benchmark `BenchmarkRunSequential_Final1M` en cinco corridas independientes con el esquema `-benchtime=1x -count=5`, lo que produce una medición por repetición y evita que la calibración automática de Go distorsione el escenario de comparación. La media de las cinco corridas fue de 2.918 s, mientras que la media recortada fue de 2.908 s, con una desviación estándar de 0.043 s y un coeficiente de variación de 1.5%. En términos de rendimiento, esto equivale a un throughput recortado aproximado de 344,119 documentos por segundo.

## Consumo de memoria

El consumo de memoria pico (peak memory) del pipeline secuencial se midió utilizando `runtime.MemStats.Sys` de Go, expresado en megabytes (MB), con el mismo criterio que el pipeline concurrente. Los resultados de las cinco corridas fueron:

- Run 1: 260.35 MB
- Run 2: 260.87 MB
- Run 3: 264.87 MB
- Run 4: 265.14 MB
- Run 5: 265.14 MB

**Media**: 263.27 MB  
**Media recortada** (eliminando mínimo y máximo): 263.63 MB  
**Rango**: 260.35 — 265.14 MB

El consumo de memoria es significativamente mayor en el secuencial que en benchmarks aislados de tokenización o lematización, reflejando el costo de mantener en memoria el corpus completo de 1,000,000 documentos. Esta métrica permite una comparación iso-funcional con la versión concurrente: ambas procesan el mismo dataset, por lo que ambas deben cargar aproximadamente el mismo volumen de datos. Variaciones en memoria entre secuencial y concurrente indicarían diferencias estructurales en cómo se asignan y liberan recursos durante el procesamiento.

El resultado es relevante porque el secuencial funciona aquí como control experimental y no como una simple ejecución auxiliar. Bajo la convención solicitada para el análisis, la razón $T_{seq} / T_{conc}$ permite evaluar si la versión concurrente realmente reduce el tiempo de ejecución respecto de la base secuencial. Con la referencia concurrente compartida por el colega, el cociente secuencial/concurrente queda por debajo de 1 tanto frente al caso concurrente de 1 worker y batch=1000 como frente a la mejor configuración reportada. En otras palabras, las mediciones actuales sugieren que el paralelismo introducido en esa implementación no compensa el costo de coordinación y sincronización, al menos para la carga y el hardware utilizados en la prueba.

Desde el punto de vista metodológico, esto obliga a ser cuidadoso con la interpretación del speedup. El pipeline secuencial aquí medido es más corto, más estable y libre de contención, por lo que debe usarse como referencia primaria del estudio. La comparación con el concurrente solo es válida si ambos modelos procesan exactamente el mismo corpus, conservan las mismas etapas lógicas y se ejecutan bajo condiciones equivalentes de hardware. En caso contrario, cualquier speedup observado puede mezclar efectos de implementación con diferencias de carga computacional.

En consecuencia, el principal aporte del secuencial no es demostrar paralelismo, sino proporcionar una línea base reproducible para el cálculo de speedup, el análisis de escalabilidad y la discusión de trade-offs, incluyendo now el consumo de memoria como dimensión adicional de comparación. La ausencia de goroutines, canales y mutex elimina costos de sincronización y simplifica la trazabilidad del flujo de documentos, lo que reduce el ruido experimental y hace más defendible la comparación posterior con el modelo concurrente. Si la versión concurrente no supera esta línea base en tiempo manteniendo un consumo de memoria comparable, el análisis debe explicarlo como un caso en el que el overhead del paralelismo domina sobre sus beneficios.
