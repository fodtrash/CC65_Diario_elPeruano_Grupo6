# Tabla Maestra de Estadísticas — Pipeline Secuencial

Esta tabla resume el procesamiento end-to-end del dataset final de 1,000,000 registros, utilizando los datos de la ejecución controlada del 11/05/2026.

## Estadísticas por configuración

| Configuración | Mediciones (s) | Media | Media recortada | CV% | Throughput recortado (docs/s) | Peak Memory (MB) |
|:---|:---|---:|---:|---:|---:|---:|
| Final 1M, 1 worker | 527.87, 538.05, 558.77, 685.16, 665.05 | 594.981 | **587.291** | 11.5% | 1,702.73 | ~285.08 |

* **Media recortada**: Se calculó eliminando el valor más bajo (527.87 s) y el más alto (685.16 s) para mitigar el ruido del sistema operativo.
* **Peak Memory**: Consumo máximo de RAM registrado durante la ejecución del benchmark.

## Resultados individuales (Actualizados)

| Run | Tiempo total (s) | Throughput (docs/s) | ns/doc | Peak Memory (MB) |
|:---:|---:|---:|---:|---:|
| 1 | 527.873 | 1,894.40 | 527,872.57 | 283.97 |
| 2 | 538.052 | 1,854.71 | 538,052.04 | 284.14 |
| 3 | 558.769 | 1,789.47 | 558,769.12 | 285.73 |
| 4 | 685.165 | 1,478.47 | 685,164.71 | 287.42 |
| 5 | 665.050 | 1,503.56 | 665,049.71 | 284.14 |

## Comparación de Speedup

Tomando como base la media recortada de **587.291 s**:

| Configuración concurrente | T_conc (s) | Speedup ($T_{seq} / T_{conc}$) |
|:---|---:|---:|
| N=1, b=1000 (concurrente) | 9.007 | 65.20x |
| N=8, b=5000 (mejor config) | 4.670 | **125.76x** |