# Resultados de benchmarks

Los archivos JSON de metricas no estan en Git (son regenerables). Este README documenta como obtenerlos.

## Estructura

```
resultados/
├── README.md                  # Este archivo
├── raw/                       # Datos crudos: 7 configs x 5 repeticiones = 35 archivos
│   ├── n1_b1000_run1.json
│   ├── n1_b1000_run2.json
│   ├── ...
│   └── n8_b5000_run5.json
└── *.json                     # Metricas sueltas de corridas ad-hoc
```

## Convencion de nombrado (raw/)

`n{workers}_b{batch}_run{i}.json`

- `n` = numero de workers (igual para clean, token, lemma)
- `b` = tamano de batch
- `run` = numero de repeticion (1-5)

## Configuraciones del benchmark

| Config | Workers | Batch | Proposito |
|---|---|---|---|
| n1_b1000 | 1 | 1000 | Baseline secuencial |
| n2_b1000 | 2 | 1000 | Escalabilidad |
| n4_b1000 | 4 | 1000 | Escalabilidad |
| n8_b1000 | 8 | 1000 | Escalabilidad (optimo) |
| n16_b1000 | 16 | 1000 | Over-subscription |
| n8_b100 | 8 | 100 | Efecto batch chico |
| n8_b5000 | 8 | 5000 | Efecto batch grande |

## Como regenerar

```powershell
.\run_benchmarks.ps1
```

Requiere el dataset completo en `data/dataset_final_1M.csv`. Genera 35 archivos en `raw/`.

## Resumen de resultados (media recortada)

Media recortada: de las 5 repeticiones, se eliminan la mas rapida y la mas lenta, y se promedian las 3 restantes.

| Config | Media recortada (ms) | Speedup vs n1 |
|---|---|---|
| n1_b1000 | 8309 | 1.00x |
| n2_b1000 | 5364 | 1.55x |
| n4_b1000 | 3724 | 2.23x |
| n8_b1000 | 3123 | **2.66x** |
| n16_b1000 | 3186 | 2.61x |
| n8_b100 | 3656 | 2.27x |
| n8_b5000 | 3417 | 2.43x |
