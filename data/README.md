# Datos del proyecto

## Estructura

- **`dataset_final_1M.csv`** — Dataset completo de 1,000,000 registros
  (244,777 reales del Diario Oficial El Peruano + 755,223 sintéticos).
  Este archivo **NO está en Git** por su tamaño (~200 MB).
  Cómo obtenerlo: ejecutar el notebook
  `notebooks/consolidacion_dataset_aumentado.ipynb` en Google Colab y
  descargar el CSV resultante a esta carpeta.

- **`sample/sample_500.csv`** — Sample de 500 registros para tests rápidos.
  Sí está en Git. Útil para correr `go test` y verificaciones básicas
  sin descargar el dataset grande.

## Columnas del dataset

| Columna            | Descripción                                      |
|--------------------|--------------------------------------------------|
| FECHA_PUBLICACION  | Fecha de publicación del dispositivo (YYYYMMDD)  |
| ENTIDAD            | Entidad pública emisora del dispositivo          |
| DISPOSITIVO        | Tipo de dispositivo legal (Resolución, Decreto…) |
| NUMERO             | Número identificador del dispositivo             |
| SUMILLA            | Título o resumen del dispositivo legal           |
| FECHA_CORTE        | Fecha de corte del registro (YYYYMMDD)           |
| ORIGEN             | Indica si el registro es REAL o SINTETICO        |

## Uso

Para correr el pipeline concurrente sobre el dataset completo:

```bash
go run ./src/concurrent/ -input data/dataset_final_1M.csv
```

Para correr sobre el sample (rápido, para tests):

```bash
go run ./src/concurrent/ -input data/sample/sample_500.csv
```
