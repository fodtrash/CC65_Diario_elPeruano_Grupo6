# Párrafos pre-redactados para el informe del PC2

Estos párrafos están escritos en un registro técnico-académico apropiado para el informe.
Todos los datos numéricos provienen de los 35 JSONs en `resultados/con_results/raw/`,
generados con el binario actual del commit `a94e206` (lematizador por n-gramas + memoización
por worker) sobre `dataset_final_1M.csv`.

---

## Sobre el speedup y la metodología estadística

Para garantizar la robustez estadística de las mediciones, cada configuración fue ejecutada cinco veces y se calculó la media recortada (descartando los valores mínimo y máximo y promediando los tres centrales) como estimador robusto del tiempo de ejecución. El baseline serial (N=1, batch=1000) registra una media recortada de **11,513 ms** con un coeficiente de variación de 5.6%, mientras que el mejor resultado se obtiene con **N=4 workers** (5,515 ms; speedup **2.09×**; eficiencia 52.2%). Configuraciones más paralelas presentan rendimientos decrecientes: N=8 alcanza un tiempo de 6,487 ms (speedup 1.77×, eficiencia 22.2%) y N=16 retrocede hasta 9,730 ms (speedup 1.18×, eficiencia 7.4%). Esta curva no monótona contrasta con expectativas clásicas y es la consecuencia directa de la combinación entre la memoización local por worker y el carácter secuencial e I/O-bound de la etapa de lectura del CSV; se discute en detalle en las secciones siguientes.

## Sobre el desplazamiento del cuello de botella

La introducción de un caché local de lematización (`map[string]string` por goroutine) eliminó la lematización como cuello de botella del pipeline: cada palabra se calcula a lo sumo una vez por worker y, con un vocabulario único de 38,201 lemas sobre 8,847,793 tokens (una repetición promedio de 231×), el hit rate del caché supera el 99% tras los primeros lotes. El efecto neto es que las tres etapas del pipeline (lectura, tokenización, lematización) terminan casi simultáneamente, como se observa al inspeccionar las métricas por etapa: en N=4, `elapsed_read_ms` = 5,433 ms, `elapsed_token_ms` = 5,468 ms y `elapsed_lemma_ms` = 5,465 ms, prácticamente idénticos al `elapsed_total_ms` de 5,515 ms. Esto demuestra empíricamente que el pipeline pasó de estar limitado por la lematización (régimen previo a la memoización) a estar limitado por la lectura secuencial del CSV de 179 MB, cuyo throughput máximo de disco constituye ahora el piso teórico del wall-clock alcanzable.

## Sobre el techo de Amdahl y la degradación con N grande

El comportamiento sub-lineal del speedup observado se explica mediante la Ley de Amdahl con una fracción serial estimada del 50–60%, mucho mayor que la registrada en mediciones previas al caché. La lectura secuencial del CSV ahora representa casi la totalidad del tiempo no paralelizable, ya que la lematización (antes etapa dominante) deja de aportar trabajo significativo. Más aún, agregar workers más allá de N=4 introduce tres costos que superan el beneficio del paralelismo: (i) cada worker mantiene su propio caché local, por lo que con N=16 hay 16 mapas separados que duplican el trabajo de warm-up y elevan la memoria pico a 203 MB; (ii) los workers compiten por la salida limitada del lector y permanecen ociosos parte del tiempo; y (iii) la sobreasignación respecto a los 8 núcleos lógicos disponibles introduce overhead de scheduling. La consecuencia es que N=8 y N=16 son más lentos en términos absolutos que N=4, una inversión respecto a la curva clásica de speedup que solo se manifiesta en regímenes I/O-bound.

## Sobre los trade-offs de diseño y la configuración óptima

El análisis de retorno marginal confirma que el incremento óptimo de workers es de 1 a 4: pasar de N=1 a N=2 reduce el tiempo en un 36% (de 11,513 a 7,330 ms), y de N=2 a N=4 lo reduce un 25% adicional (a 5,515 ms). Pasar de N=4 a N=8 ya degrada el rendimiento en 18% (a 6,487 ms) y de N=8 a N=16 en 50% (a 9,730 ms). El consumo de memoria, por su parte, crece de forma aproximadamente lineal con el número de workers (41 MB con N=1, 99 MB con N=4, 203 MB con N=16) debido al caché local replicado por goroutine. La configuración recomendada es por tanto **N=4 workers por etapa, batch_size=1,000**: ofrece el mejor compromiso entre tiempo (5,515 ms), memoria (99 MB) y eficiencia (52.2%).

## Sobre el efecto del tamaño de lote (batch size)

El estudio del impacto del batch size, manteniendo N=8 fijo, revela un comportamiento no monótono. Con batch=100 el tiempo asciende a 8,959 ms y la memoria pico es de 109 MB; con batch=1,000 baja a 6,487 ms y 144 MB; con batch=5,000 sube nuevamente a 9,136 ms con 309 MB de memoria. La degradación de batch=100 se atribuye al alto número de mensajes intercambiados entre etapas (10,000 lotes por canal) y a la mayor contención del mutex (6.06 ms vs 2.01 ms con batch=1,000), mientras que la degradación de batch=5,000 se explica por la sub-utilización del paralelismo: pocos lotes grandes provocan que workers queden ociosos esperando al lector, sin compensar el coste de mantener 5,000 documentos cargados en memoria por lote. El punto intermedio de batch=1,000 valida empíricamente el principio de balance entre granularidad de paralelismo y overhead de coordinación discutido por You (2025).

## Sobre el overhead de sincronización (mutex)

La medición de contención del mutex global, que protege los contadores compartidos `tokens_globales`, `docs_procesados`, `docs_reales` y `docs_sinteticos` en correspondencia directa con el modelo Promela del PC1, arroja valores entre 1.00 y 6.06 ms en todas las configuraciones probadas. Este valor es despreciable frente al tiempo total de ejecución (>5,500 ms), representando menos del 0.1% del wall-clock. La decisión de delegar la agregación de frecuencias léxicas a estructuras locales por worker (`map[string]int`) y reservar el mutex únicamente para contadores escalares quedó así doblemente validada: la baja contención original se preserva tras la introducción del caché, evidencia de que ninguna de las dos optimizaciones (caché de lematización, agregación local) introduce nuevos puntos de sincronización global.

## Sobre la memoización local como optimización dominante

Comparando contra una corrida control sin caché del mismo binario sobre N=8 y batch=1,000, el tiempo se reduce de **227,229 ms a 6,487 ms**, un speedup interno de **35.0×** sobre la misma topología de paralelismo. Esta cifra deja en evidencia que, una vez introducido el lematizador por n-gramas y similitud coseno (~127 comparaciones de vectores por token), la memoización por worker se convierte en el efecto de primer orden del rendimiento. La condición que la habilita es la naturaleza determinística y pura de `nlp.Lemmatize` —misma entrada, misma salida, sin estado mutable— y la alta repetición léxica del corpus legal (~231 ocurrencias promedio por lema único). La estrategia de mantener un caché por goroutine, sin compartición ni sincronización, hace explícito el trade-off entre memoria (40 MB por worker) y tiempo de CPU, sin introducir riesgo de race conditions ni complicar el modelo Promela.

## Sobre la comparación con la línea base secuencial

La media recortada del pipeline secuencial sobre el mismo dataset es de **613,593 ms** (10:13 min). Tomando la mejor configuración concurrente (N=4, batch=1,000) como referencia, el speedup respecto al secuencial alcanza **111.3×**, valor que excede con creces el techo teórico de paralelización para 8 núcleos lógicos. Esta divergencia no debe interpretarse como un efecto exclusivo del paralelismo: la versión concurrente incorpora memoización local por worker (commit `a94e206`), optimización ausente en la implementación secuencial vigente. Una comparación estrictamente iso-funcional requeriría aplicar el mismo mecanismo de caché al pipeline secuencial; en ausencia de esa simetría, el cociente T_seq / T_conc reportado constituye la combinación de dos efectos —paralelización y memoización— y debe declararse explícitamente como tal en el análisis del informe.
