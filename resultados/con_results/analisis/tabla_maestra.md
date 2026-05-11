# Tabla Maestra de Estadisticas - Pipeline Concurrente NLP

Datos generados con el script actual de benchmarks.
7 configuraciones x 5 repeticiones = 35 corridas sobre `data/dataset_final_1M.csv` (1,000,000 registros).
Hardware: 8 nucleos logicos, Windows 11. Metrica reportada: `elapsed_total_ms`.

## Estadisticas por configuracion

| Config | Mediciones (ms) | Media | Media recortada | StdDev | CV% |
|:---|:---|---:|---:|---:|---:|
| N=1, b=1000 | 5375, 5404, 5447, 5451, 5489 | 5,433 | **5,434** | 40 | 0.7% |
| N=2, b=1000 | 3093, 3103, 3104, 3173, 3176 | 3,130 | **3,127** | 38 | 1.2% |
| N=4, b=1000 | 2353, 2366, 2368, 2396, 2431 | 2,383 | **2,377** | 29 | 1.2% |
| N=8, b=1000 | 2344, 2411, 2511, 2560, 2614 | 2,488 | **2,494** | 99 | 4.0% |
| N=16, b=1000 | 2909, 2984, 2994, 3057, 3798 | 3,148 | **3,012** | 330 | 10.5% |
| N=8, b=100 | 2383, 2398, 2449, 2460, 2493 | 2,437 | **2,436** | 41 | 1.7% |
| N=8, b=5000 | 2333, 2363, 2375, 2387, 2389 | 2,369 | **2,375** | 20 | 0.9% |

Media recortada: se eliminan el valor minimo y el maximo de las 5 mediciones y se promedian los 3 centrales.

## Speedup con media recortada (baseline = N=1, b=1000)

| N workers | Tiempo recortado (ms) | Speedup | Eficiencia |
|:---:|---:|---:|---:|
| 1 | 5,434 | 1.00x | 100.0% |
| 2 | 3,127 | 1.74x | 86.9% |
| **4** | **2,377** | **2.29x** | **57.2%** |
| 8 | 2,494 | 2.18x | 27.2% |
| 16 | 3,012 | 1.80x | 11.3% |

Hallazgo clave con `batch=1000`: el mejor punto es N=4.

## Tiempos por etapa (media recortada, ms)

| Config | Lectura | Tokenizacion | Lematizacion | Total |
|:---|---:|---:|---:|---:|
| N=1, b=1000 | 5,418 | 5,420 | 5,419 | 5,434 |
| N=2, b=1000 | 3,115 | 3,118 | 3,114 | 3,127 |
| N=4, b=1000 | 2,371 | 2,374 | 2,365 | 2,377 |
| N=8, b=1000 | 2,476 | 2,479 | 2,460 | 2,494 |
| N=16, b=1000 | 2,999 | 3,002 | 2,972 | 3,012 |

`elapsed_read_ms`, `elapsed_token_ms` y `elapsed_lemma_ms` muestran que las tres etapas permanecen casi solapadas; la duracion total sigue de cerca el tramo mas lento del pipeline en cada configuracion.

## Analisis del batch size (N=8 fijo)

| Batch size | Tiempo recortado (ms) | Memoria pico promedio (MB) |
|:---:|---:|---:|
| 100 | 2,436 | 259 |
| 1000 | 2,494 | 258 |
| **5000** | **2,375** | **321** |

En esta corrida, `batch=5000` logra el mejor tiempo para N=8, pero con una penalizacion clara de memoria.

## Memoria pico por configuracion (promedio de las 5 reps)

| Config | Memoria pico (MB) |
|:---|---:|
| N=1, b=1000 | 227 |
| N=2, b=1000 | 242 |
| N=4, b=1000 | 251 |
| N=8, b=1000 | 258 |
| N=16, b=1000 | 251 |
| N=8, b=100 | 259 |
| N=8, b=5000 | 321 |

La memoria aumenta con lotes mas grandes y, en general, crece al subir el paralelismo hasta N=8 en este entorno.

## Contencion del mutex

| Config | mutex_contention_ms (max de 5 reps) |
|:---|---:|
| N=1, b=1000 | 0.00 |
| N=2, b=1000 | 0.00 |
| N=4, b=1000 | 0.00 |
| N=8, b=1000 | 0.00 |
| N=16, b=1000 | 0.00 |
| N=8, b=100 | 0.00 |
| N=8, b=5000 | 0.00 |

No se observa contencion medible del lock en estas 35 corridas.

## Consistencia de salida

Se verifico determinismo en las 35 ejecuciones:
- `total_tokens = 8,847,717`
- `total_lemmas_unique = 38,201`
- `docs_reales = 244,779`
- `docs_sinteticos = 755,221`

Los valores se mantienen identicos sin importar N o batch.

## Comparacion con la linea base secuencial

La media recortada secuencial reportada es **613,593 ms** (~10:13 min) segun `resultados/seq_results/resumen_benchmarks_seq.md`.

| Config concurrente | Tiempo conc. (ms) | T_seq / T_conc |
|:---|---:|---:|
| N=1, b=1000 | 5,434 | 113.0x |
| N=2, b=1000 | 3,127 | 196.2x |
| **N=4, b=1000** | **2,377** | **258.1x** |
| N=8, b=1000 | 2,494 | 246.0x |
| N=16, b=1000 | 3,012 | 203.7x |

Nota metodologica: esta comparacion combina paralelizacion y optimizaciones internas del pipeline concurrente; no es una comparacion iso-funcional pura si el secuencial no replica exactamente las mismas optimizaciones.

## Hallazgos clave

1. Con `batch=1000`, el mejor speedup aparece en N=4 (2.29x), con buena eficiencia (57.2%).
2. N=16 muestra alta variabilidad (CV 10.5%) por un outlier de 3,798 ms.
3. Para N=8, `batch=5000` mejora tiempo frente a `batch=1000`, pero sube memoria de ~258 MB a 321 MB.
4. La contencion de mutex no es factor limitante en esta tanda (0.00 ms max por configuracion).
5. Los resultados funcionales son deterministas en todas las corridas (tokens, lemas y conteo real/sintetico constantes).
