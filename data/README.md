# Data

## Dataset completo

El dataset completo (~1,000,000 registros) **no se incluye en este repositorio** debido a su tamaño.

## ¿Cómo obtenerlo?

El dataset se genera ejecutando el notebook de limpieza y aumentación de datos ubicado en:

```
notebooks/consolidacion_dataset_aumentado.ipynb
```

### Pasos

1. Asegúrate de tener **Python 3** y **Jupyter Notebook** instalados.
2. Instala las dependencias necesarias:
   ```bash
   pip install pandas numpy matplotlib jupyter
   ```
3. Abre el notebook:
   ```bash
   jupyter notebook notebooks/consolidacion_dataset_aumentado.ipynb
   ```
4. Ejecuta todas las celdas en orden (**Run All**).
5. Al finalizar, se generará el archivo `dataset_final_1M.csv` con aproximadamente **1,000,000 de registros**.

## Muestra del dataset

En la carpeta `data/sample/` se incluye una muestra pequeña del dataset para pruebas rápidas sin necesidad de generar el completo.

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
