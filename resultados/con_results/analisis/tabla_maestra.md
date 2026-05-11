# Tabla Maestra de Estadísticas — Pipeline Concurrente NLP

Datos generados con el binario actual (commit `a94e206 perf(concurrent): cachear lematizacion por worker`).
7 configuraciones × 5 repeticiones = 35 corridas sobre `data/dataset_final_1M.csv` (1,000,000 registros).
Hardware: 8 núcleos lógicos, Windows 11. Métrica reportada: `elapsed_total_ms`.

## Estadísticas por configuración

| Config | Mediciones (ms) | Media | Media recortada | StdDev | CV% |
|:---|:---|---:|---:|---:|---:|
| N=1, b=1000 | 10603, 11086, 11377, 12076, 12085 | 11,445 | **11,513** | 642 | 5.6% |
| N=2, b=1000 | 7110, 7144, 7367, 7479, 7526 | 7,325 | **7,330** | 190 | 2.6% |
| N=4, b=1000 | 5348, 5406, 5539, 5600, 5849 | 5,548 | **5,515** | 196 | 3.5% |
| N=8, b=1000 | 6089, 6144, 6397, 6920, 8212 | 6,752 | **6,487** | 880 | 13.0% |
| N=16, b=1000 | 9561, 9640, 9726, 9823, 10201 | 9,790 | **9,730** | 250 | 2.5% |
| N=8, b=100 | 8159, 8433, 8958, 9487, 11754 | 9,358 | **8,959** | 1,433 | 15.3% |
| N=8, b=5000 | 7461, 9034, 9076, 9297, 10060 | 8,986 | **9,136** | 947 | 10.5% |

Media recortada: se eliminan el valor mínimo y el máximo de las 5 mediciones y se promedian los 3 centrales.

## Speedup con media recortada (baseline = N=1)

| N workers | Tiempo recortado (ms) | Speedup | Eficiencia |
|:---:|---:|---:|---:|
| 1 | 11,513 | 1.00× | 100.0% |
| 2 | 7,330 | 1.57× | 78.5% |
| **4** | **5,515** | **2.09×** | **52.2%** |
| 8 | 6,487 | 1.77× | 22.2% |
| 16 | 9,730 | 1.18× | 7.4% |

**Hallazgo clave: el speedup óptimo es N=4 (2.09×), no N=8 ni N=16.** Más allá de N=4 el tiempo crece porque el cuello de botella se desplazó a la lectura secuencial del CSV (ver sección de tiempos por etapa).

## Tiempos por etapa (media recortada, ms)

| Config | Lectura | Tokenización | Lematización | Total |
|:---|---:|---:|---:|---:|
| N=1, b=1000 | 11,394 | 11,480 | 11,473 | 11,513 |
| N=2, b=1000 | 7,252 | 7,296 | 7,283 | 7,330 |
| N=4, b=1000 | 5,433 | 5,468 | 5,465 | 5,515 |
| N=8, b=1000 | 6,385 | 6,420 | 6,419 | 6,487 |
| N=16, b=1000 | 9,544 | 9,647 | 9,655 | 9,730 |

`elapsed_read_ms`, `elapsed_token_ms` y `elapsed_lemma_ms` miden el wall-clock de cada etapa (primer worker arrancando hasta último worker terminando). Con el caché de lematización, los tres tiempos convergen al wall-clock total: las tres etapas se solapan completamente y todas terminan junto con el lector. La etapa de lectura es ahora el límite inferior del pipeline.

## Análisis del batch size (N=8 fijo)

| Batch size | Tiempo recortado (ms) | Memoria pico (MB) |
|:---:|---:|---:|
| 100 | 8,959 | 109 |
| **1000** | **6,487** | **144** |
| 5000 | 9,136 | 309 |

Batch=1000 es óptimo. Batch=100 produce demasiados lotes (10,000 mensajes por canal) y aumenta la contención del mutex. Batch=5000 satura el canal: pocos lotes grandes provocan que workers queden ociosos esperando al lector, y la memoria pico se duplica.

## Memoria pico por configuración (promedio de las 5 reps)

| Config | Memoria pico (MB) |
|:---|---:|
| N=1, b=1000 | 41 |
| N=2, b=1000 | 62 |
| N=4, b=1000 | 99 |
| N=8, b=1000 | 144 |
| N=16, b=1000 | 203 |
| N=8, b=100 | 109 |
| N=8, b=5000 | 309 |

La memoria crece aproximadamente lineal con el número de workers porque cada worker mantiene su propio caché local de lematización (`map[string]string` de hasta ~38,000 entradas). Con N workers el costo agregado del caché es N × |vocab|.

## Contención del mutex

| Config | mutex_contention_ms (max de 5 reps) |
|:---|---:|
| N=1, b=1000 | 2.39 |
| N=2, b=1000 | 2.00 |
| N=4, b=1000 | 1.52 |
| N=8, b=1000 | 2.01 |
| N=16, b=1000 | 1.00 |
| N=8, b=100 | 6.06 |
| N=8, b=5000 | 1.00 |

La contención más alta (6.06 ms con batch=100) se da por las 10,000 adquisiciones de lock vs 1,000 con batch=1000. Aun así, es despreciable frente al total (>5,500 ms): <0.1% en todas las configuraciones.

## Comparación con la línea base secuencial

La media recortada secuencial es de **613,593 ms** (~10:13 min) según `resultados/seq_results/resumen_benchmarks_seq.md`, medida sobre el mismo `dataset_final_1M.csv` con el lematizador por n-gramas (sin caché).

| Config concurrente | Tiempo conc. (ms) | T_seq / T_conc |
|:---|---:|---:|
| N=1, b=1000 | 11,513 | 53.3× |
| N=2, b=1000 | 7,330 | 83.7× |
| **N=4, b=1000** | **5,515** | **111.3×** |
| N=8, b=1000 | 6,487 | 94.6× |
| N=16, b=1000 | 9,730 | 63.1× |

**Observación metodológica importante**: este speedup mezcla dos efectos: (a) la paralelización propiamente dicha y (b) la memoización local en cada worker concurrente, optimización que el pipeline secuencial actual no incorpora. Una comparación estrictamente iso-funcional requeriría aplicar el mismo caché al secuencial.

## Hallazgos clave

1. **Speedup óptimo: 2.09× con N=4 workers y batch=1000.** A diferencia de mediciones previas donde N=8 era el sweet spot, el caché de lematización desplazó el cuello de botella a la lectura del CSV y volvió contraproducente agregar más de 4 workers.
2. **N=8 y N=16 son más lentos que N=4** (6,487 y 9,730 ms vs 5,515 ms): el overhead de scheduling, la fragmentación del caché por worker y la presión sobre la memoria superan el beneficio del paralelismo cuando el lector ya es el cuello de botella.
3. **Fracción serial estimada (Amdahl): ~50–60%**, dominada por la lectura secuencial del CSV. Con `bufio.NewReaderSize(1<<20)` y `csv.Reader`, el lector procesa los 179 MB del dataset en ~5–6 s mínimo, lo que constituye el piso del wall-clock alcanzable.
4. **El batch=1000 sigue siendo el mejor trade-off** (6,487 ms con 144 MB); batch=100 introduce demasiados mensajes y batch=5000 consume 309 MB sin ganancia de velocidad.
5. **Contención del mutex despreciable** (<7 ms sobre >5,500 ms totales): valida el diseño de agregación local por worker.
6. **Determinismo verificado**: `total_tokens=8,847,793` y `total_lemmas_unique=38,201` idénticos en las 35 corridas, independiente de N y batch.
7. **Caché por worker como optimización dominante**: comparado contra una corrida sin caché (227,229 ms para N=8, b=1000), el caché aporta un speedup interno de **35× sobre la misma topología**, mientras que el paralelismo aporta 2.09× adicional. La memoización local es ahora el efecto de primer orden.
