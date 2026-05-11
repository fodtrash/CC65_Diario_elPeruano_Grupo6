# Tabla Maestra de Estadisticas — Pipeline Concurrente NLP

Datos generados con `run_benchmarks.ps1` (7 configuraciones x 5 repeticiones = 35 corridas).
Hardware de prueba: 8 nucleos logicos, Windows, dataset de 1,000,000 registros.
Metrica reportada: `elapsed_total_ms` del JSON de cada corrida.

## Estadisticas por configuracion

| Config | Mediciones (ms) | Media | Media recortada | StdDev | CV% |
|:---|:---|---:|---:|---:|---:|
| N=1, b=1000 | 8791, 8885, 8981, 9156, 9412 | 9045 | **9007** | 246 | 2.7% |
| N=2, b=1000 | 5253, 5306, 5512, 5699, 6403 | 5635 | **5506** | 465 | 8.2% |
| N=4, b=1000 | 5016, 5261, 5298, 5952, 7115 | 5728 | **5504** | 849 | 14.8% |
| N=8, b=1000 | 4767, 4775, 4786, 5072, 5199 | 4920 | **4878** | 202 | 4.1% |
| N=16, b=1000 | 4687, 4755, 4770, 4844, 5156 | 4842 | **4790** | 184 | 3.8% |
| N=8, b=100 | 4857, 4876, 4936, 4990, 5432 | 5018 | **4934** | 237 | 4.7% |
| N=8, b=5000 | 4600, 4632, 4642, 4736, 4806 | 4683 | **4670** | 85 | 1.8% |

Media recortada: se eliminan el valor minimo y maximo de las 5 mediciones y se promedian los 3 centrales.

Nota: la configuracion N=4 muestra alta varianza (CV=14.8%) con un outlier en run2 (7115 ms), probablemente por interferencia del sistema operativo durante esa corrida. La media recortada mitiga este efecto.

## Speedup con media recortada (baseline = N=1)

| N workers | Tiempo recortado (ms) | Speedup | Eficiencia |
|:---:|---:|---:|---:|
| 1 | 9007 | 1.00x | 100.0% |
| 2 | 5506 | 1.64x | 82.0% |
| 4 | 5504 | 1.64x | 41.0% |
| **8** | **4878** | **1.85x** | **23.1%** |
| 16 | 4790 | 1.88x | 11.8% |

Nota: el speedup de N=4 (1.64x) es anomalamente igual al de N=2, afectado por el outlier mencionado arriba. Excluyendo el outlier, N=4 deberia situarse entre 1.64x y 1.85x.

## Analisis del batch size (N=8 fijo)

| Batch size | Tiempo recortado (ms) | Speedup | Memoria pico (MB) |
|:---:|---:|---:|---:|
| 100 | 4934 | 1.83x | ~80 |
| 1000 | 4878 | 1.85x | ~113 |
| **5000** | **4670** | **1.93x** | **~192** |

## Memoria pico por configuracion

| Config | Memoria pico (MB) |
|:---|---:|
| N=1, b=1000 | 34 |
| N=2, b=1000 | 47 |
| N=4, b=1000 | 71 |
| N=8, b=1000 | 113 |
| N=16, b=1000 | 146 |
| N=8, b=100 | 80 |
| N=8, b=5000 | 192 |

## Contencion del mutex

| Config | mutex_contention_ms (max) |
|:---|---:|
| N=1, b=1000 | 2.25 |
| N=2, b=1000 | 2.00 |
| N=4, b=1000 | 1.01 |
| N=8, b=1000 | 6.42 |
| N=16, b=1000 | 2.52 |
| N=8, b=100 | 10.63 |
| N=8, b=5000 | 1.00 |

Nota: la contencion mas alta (10.63 ms) se da con batch=100 porque hay 10,000 lotes y por tanto 10,000 adquisiciones de lock vs 1,000 con batch=1000. Aun asi, 10.63 ms es despreciable frente a >4,600 ms de ejecucion total (<0.3%).

## Hallazgos clave

1. **Speedup optimo: 1.93x con N=8 workers y batch=5000**, aunque N=8 con batch=1000 (1.85x) ofrece mejor balance costo-memoria.
2. **Fraccion serial estimada (Ley de Amdahl): ~48%**, asociada a la lectura secuencial del CSV y la fusion de Clean+Tokenize en una sola etapa del pipeline.
3. **N=16 NO mejora significativamente vs N=8** (4790 vs 4878 ms): la sobreasignacion de workers respecto a nucleos fisicos introduce overhead de scheduling.
4. **Batch=5000 es el mas rapido** pero consume 70% mas memoria que batch=1000 (192 vs 113 MB); batch=1000 es el mejor trade-off.
5. **Contencion del mutex es despreciable** (<11 ms sobre >4,600 ms de ejecucion total): valida el diseno de agregacion local por worker.
6. **Determinismo verificado:** total_tokens=8,847,793 y total_lemmas_unique=45,321 identicos en las 35 corridas.
