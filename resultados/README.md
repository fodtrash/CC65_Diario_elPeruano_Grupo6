# Resultados Experimentales — PC2 Pipeline NLP

Esta carpeta agrupa los resultados de los dos pipelines del trabajo:
el concurrente (con goroutines y canales) y el secuencial (línea base).
Las dos sub-carpetas comparten dataset (`data/dataset_final_1M.csv`,
1,000,000 registros) para que las mediciones sean comparables.

## Estructura

```
resultados/
├── README.md                # Este índice
├── con_results/             # Pipeline concurrente (rama feature/concurrent)
│   ├── README.md
│   ├── raw/                 # 35 JSONs (7 configs × 5 reps)
│   └── analisis/
│       ├── tabla_maestra.md
│       └── parrafos_informe.md
└── seq_results/             # Pipeline secuencial (línea base)
    ├── raw/                 # 5 JSONs del benchmark go test
    ├── resumen_benchmarks_seq.md
    └── analisis_benchmark_seq.md
```

## Comparación general

| Pipeline | Configuración | Tiempo recortado (1M docs) | Throughput |
|:---|:---|---:|---:|
| Secuencial | 1 worker (`BenchmarkRunSequential_Final1M`) | **613,593 ms** (~10:13 min) | 1,629 docs/s |
| Concurrente | N=4 workers, batch=1,000 | **5,515 ms** (~5.5 s) | 181,323 docs/s |
| Speedup conc. vs seq. | | **111.3×** | |

Detalle por configuración en [`con_results/analisis/tabla_maestra.md`](con_results/analisis/tabla_maestra.md) y métodos en cada README sub-carpeta.

## Nota sobre la comparación

El speedup reportado (~111×) excede el techo teórico de paralelización para 8 núcleos. Esto se debe a que el pipeline concurrente incorpora **memoización local por worker en la lematización** (commit `a94e206`), optimización que el pipeline secuencial actual no aplica. Una comparación estrictamente iso-funcional exigiría aplicar el mismo caché al secuencial; en ausencia de esa simetría, el cociente T_seq / T_conc representa la combinación de dos efectos: paralelización y memoización. Esta limitación está documentada explícitamente en [`con_results/analisis/parrafos_informe.md`](con_results/analisis/parrafos_informe.md) y en [`seq_results/analisis_benchmark_seq.md`](seq_results/analisis_benchmark_seq.md).
