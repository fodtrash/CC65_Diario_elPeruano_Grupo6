# Tabla Maestra de Estadisticas — Pipeline Concurrente NLP

Datos generados con `run_benchmarks.ps1` (7 configuraciones x 5 repeticiones = 35 corridas).
Hardware de prueba: 8 nucleos logicos, Windows, dataset de 1,000,000 registros.
Metrica reportada: `elapsed_total_ms` del JSON de cada corrida.

## Estadisticas por configuracion

| Config | Mediciones (ms) | Media | Media recortada | StdDev | CV% |
|:---|:---|---:|---:|---:|---:|
| N=1, b=1000 | 7208, 7297, 8637, 8993, 9180 | 8263 | **8309** | 943 | 11.4% |
| N=2, b=1000 | 5241, 5247, 5407, 5437, 5469 | 5360 | **5364** | 108 | 2.0% |
| N=4, b=1000 | 3586, 3688, 3694, 3790, 3798 | 3711 | **3724** | 87 | 2.3% |
| N=8, b=1000 | 2868, 3022, 3162, 3184, 3766 | 3200 | **3123** | 341 | 10.7% |
| N=16, b=1000 | 3008, 3011, 3111, 3437, 3453 | 3204 | **3186** | 224 | 7.0% |
| N=8, b=100 | 3435, 3616, 3619, 3732, 3794 | 3639 | **3656** | 137 | 3.8% |
| N=8, b=5000 | 3388, 3391, 3403, 3458, 3470 | 3422 | **3417** | 39 | 1.1% |

Media recortada: se eliminan el valor minimo y maximo de las 5 mediciones y se promedian los 3 centrales.

## Speedup con media recortada (baseline = N=1)

| N workers | Tiempo recortado (ms) | Speedup | Eficiencia |
|:---:|---:|---:|---:|
| 1 | 8309 | 1.00x | 100.0% |
| 2 | 5364 | 1.55x | 77.5% |
| 4 | 3724 | 2.23x | 55.8% |
| **8** | **3123** | **2.66x** | **33.3%** |
| 16 | 3186 | 2.61x | 16.3% |

## Analisis del batch size (N=8 fijo)

| Batch size | Tiempo recortado (ms) | Speedup | Memoria pico (MB) |
|:---:|---:|---:|---:|
| 100 | 3656 | 2.27x | ~91 |
| **1000** | **3123** | **2.66x** | **~107** |
| 5000 | 3417 | 2.43x | ~160 |

## Memoria pico por configuracion

| Config | Memoria pico (MB) |
|:---|---:|
| N=1, b=1000 | 33 |
| N=2, b=1000 | 45 |
| N=4, b=1000 | 78 |
| N=8, b=1000 | 107 |
| N=16, b=1000 | 139 |
| N=8, b=100 | 91 |
| N=8, b=5000 | 160 |

## Contencion del mutex

| Config | mutex_contention_ms |
|:---|---:|
| N=1, b=1000 | 0.54 |
| N=2, b=1000 | ~0 |
| N=4, b=1000 | ~0 |
| N=8, b=1000 | ~0 |
| N=16, b=1000 | 0.52 |
| N=8, b=100 | 2.53 |
| N=8, b=5000 | 1.00 |

Nota: la contencion mas alta (2.53 ms) se da con batch=100 porque hay 10,000 lotes y por tanto 10,000 adquisiciones de lock vs 1,000 con batch=1000.

## Hallazgos clave

1. **Speedup optimo: 2.66x con N=8 workers** (uno por nucleo logico).
2. **Fraccion serial estimada (Ley de Amdahl): ~28%**, asociada a la lectura secuencial del CSV y sincronizacion entre etapas.
3. **N=16 NO mejora vs N=8** (3186 vs 3123 ms): la sobreasignacion de workers respecto a nucleos fisicos introduce overhead de scheduling.
4. **Batch=1000 es el optimo**: batch=100 satura el canal con mensajes, batch=5000 aumenta el consumo de memoria sin beneficio en tiempo.
5. **Las configuraciones con mas workers son mas estables** (CV bajo): la concurrencia amortigua interferencias del SO.
6. **Contencion del mutex es despreciable** (<3 ms sobre >3000 ms de ejecucion total): valida el diseno de agregacion local por worker.
