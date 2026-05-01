# Parrafos pre-redactados para el informe del PC2

Estos parrafos estan escritos en un registro tecnico-academico apropiado
para el informe. Ricardo puede usarlos literalmente o adaptarlos segun
su estilo. Todos los datos numericos provienen de los 35 JSONs en
`resultados/raw/`.

---

## Sobre el speedup y la metodologia estadistica

Para garantizar la robustez estadistica de las mediciones, cada
configuracion fue ejecutada 5 veces y se calculo la media recortada
(descartando los valores minimo y maximo de las mediciones individuales)
sobre los tres valores centrales. Esta tecnica reduce el sesgo introducido
por outliers temporales asociados a la actividad del sistema operativo
durante las pruebas. Los resultados muestran un speedup creciente desde
1.00x (N=1, baseline = 8,309 ms) hasta un maximo de 2.66x (N=8, 3,123 ms),
con un coeficiente de variacion entre 1.1% y 11.4% segun la configuracion.
La eficiencia paralela decae de 77.5% (N=2) a 33.3% (N=8) y 16.3% (N=16),
lo que evidencia el comportamiento sub-lineal predicho por la Ley de
Amdahl. Notablemente, las configuraciones con mayor numero de workers
presentan mediciones mas estables que el baseline secuencial, lo que
sugiere que la concurrencia no solo mejora el throughput sino tambien
reduce la variabilidad del rendimiento al amortizar interferencias del
sistema entre multiples nucleos.

## Sobre la scalability y el techo de Amdahl

Las pruebas con N=16 workers (sobreasignacion respecto a los 8 nucleos
logicos disponibles) muestran un retroceso del rendimiento: el tiempo
total aumenta a 3,186 ms (vs 3,123 ms con N=8) y el consumo de memoria
crece a 139 MB (vs 107 MB con N=8). Este resultado confirma empiricamente
que el sweet spot de paralelismo esta acotado por el numero de nucleos
del hardware: agregar mas goroutines mas alla del paralelismo real
disponible introduce overhead de scheduling del runtime de Go que supera
el beneficio de la concurrencia adicional. Combinado con la fraccion
serial estimada del 28% (Ley de Amdahl), esto define el techo practico
de speedup en aproximadamente 2.7x para esta arquitectura de hardware.

## Sobre los trade-offs de diseno

El analisis de retorno marginal muestra que el incremento de workers de
1 a 2 reduce el tiempo de ejecucion en un 35% (de 8,309 a 5,364 ms),
mientras que el incremento de 4 a 8 workers solo lo reduce en un 16%
adicional (de 3,724 a 3,123 ms), a costa de un aumento del 37% en
consumo de memoria (de 78 MB a 107 MB). Esto sugiere que el punto
optimo costo-beneficio se encuentra en N=4 cuando los recursos de
memoria son limitados, mientras que N=8 sigue siendo la configuracion
mas rapida en terminos absolutos cuando la memoria no es restriccion.

## Sobre el efecto del tamano de lote (batch size)

El estudio del impacto del tamano de lote sobre el rendimiento,
manteniendo N=8 fijo, revela un comportamiento en U invertida: con
batches pequenos (100 documentos) el tiempo total es 3,656 ms y la
contencion del mutex se eleva a 2.53 ms debido al alto numero de
mensajes intercambiados entre etapas (10,000 batches en total). Con
batches grandes (5,000 documentos) el tiempo sube a 3,417 ms y la
memoria pico crece a 160 MB por la carga simultanea de mas documentos
en RAM. El optimo se ubica en batch=1,000 (3,123 ms, 107 MB), validando
empiricamente la observacion de You (2025) sobre el balance entre
granularidad de paralelismo y overhead de coordinacion en pipelines de
NLP.

## Sobre el overhead de sincronizacion (mutex)

La medicion de contencion del mutex global, que protege los contadores
compartidos (tokens_globales, docs_procesados, docs_reales,
docs_sinteticos) en correspondencia directa con el modelo Promela del
PC1, arroja valores entre 0 y 2.53 ms acumulados en todas las
configuraciones probadas. Este valor despreciable frente al tiempo total
de ejecucion (>3,000 ms) valida empiricamente la decision de diseno de
delegar la agregacion de frecuencias lexicas a estructuras locales por
worker (map[string]int) y reservar el mutex unicamente para contadores
escalares, evitando el cuello de botella de sincronizacion reportado
por Treviso et al. (2023) en pipelines de NLP con estado global
compartido. La mayor contencion (2.53 ms) se observa con batch=100,
donde el numero de adquisiciones de lock es 10x mayor que con batch=1,000.

## Sobre la configuracion optima recomendada

Configuracion optima del pipeline concurrente: N=8 workers por etapa
(uno por nucleo logico) y batch_size=1,000 documentos. Esta combinacion
minimiza el tiempo total (3,123 ms para 1M registros, throughput de
~320,000 docs/s) manteniendo el consumo de memoria controlado (107 MB)
y la contencion de sincronizacion despreciable (<1 ms).
