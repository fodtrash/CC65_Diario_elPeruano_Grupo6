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
1.00x (N=1, baseline = 9,007 ms) hasta un maximo de 1.93x (N=8 con
batch=5000, 4,670 ms), con un coeficiente de variacion entre 1.8% y
14.8% segun la configuracion. La eficiencia paralela decae de 82.0%
(N=2) a 23.1% (N=8) y 11.8% (N=16), lo que evidencia el comportamiento
sub-lineal predicho por la Ley de Amdahl. Las configuraciones con mayor
numero de workers presentan mediciones mas estables (CV < 5%) que las
de menor paralelismo, lo que sugiere que la concurrencia reduce la
variabilidad del rendimiento al amortizar interferencias del sistema
entre multiples nucleos.

## Sobre la scalability y el techo de Amdahl

Las pruebas con N=16 workers (sobreasignacion respecto a los 8 nucleos
logicos disponibles) muestran un rendimiento practicamente identico al
de N=8: 4,790 ms vs 4,878 ms respectivamente (diferencia < 2%). El
consumo de memoria, en cambio, crece de 113 MB a 146 MB. Este resultado
confirma empiricamente que el sweet spot de paralelismo esta acotado
por el numero de nucleos del hardware: agregar mas goroutines mas alla
del paralelismo real disponible no aporta beneficio medible pero si
incrementa el overhead de memoria y scheduling del runtime de Go. La
fraccion serial estimada es de aproximadamente 48% (Ley de Amdahl), lo
que refleja tanto la lectura secuencial del CSV como la naturaleza del
pipeline de 3 etapas, donde la etapa de Tokenizacion integra la
limpieza de texto dentro del mismo worker.

## Sobre los trade-offs de diseno

El analisis de retorno marginal muestra que el incremento de workers de
1 a 2 reduce el tiempo de ejecucion en un 39% (de 9,007 a 5,506 ms),
mientras que el incremento de 2 a 8 workers solo lo reduce en un 11%
adicional (de 5,506 a 4,878 ms), a costa de un aumento del 140% en
consumo de memoria (de 47 MB a 113 MB). Esto sugiere que el punto
optimo costo-beneficio se encuentra en N=2 cuando los recursos de
memoria son limitados, mientras que N=8 sigue siendo la configuracion
mas rapida en terminos absolutos cuando la memoria no es restriccion.

## Sobre el efecto del tamano de lote (batch size)

El estudio del impacto del tamano de lote sobre el rendimiento,
manteniendo N=8 fijo, revela que batches mas grandes mejoran el
tiempo total: con batches pequenos (100 documentos) el tiempo es
4,934 ms y la contencion del mutex se eleva a 10.63 ms debido al
alto numero de mensajes intercambiados entre etapas (10,000 batches
en total). Con batches de 5,000 documentos el tiempo baja a 4,670 ms
pero la memoria pico crece a 192 MB por la carga simultanea de mas
documentos en RAM. El punto intermedio de batch=1,000 (4,878 ms, 113 MB)
ofrece el mejor trade-off entre rendimiento y consumo de recursos,
validando empiricamente la observacion de You (2025) sobre el balance
entre granularidad de paralelismo y overhead de coordinacion en
pipelines de NLP.

## Sobre el overhead de sincronizacion (mutex)

La medicion de contencion del mutex global, que protege los contadores
compartidos (tokens_globales, docs_procesados, docs_reales,
docs_sinteticos) en correspondencia directa con el modelo Promela del
PC1, arroja valores entre 1 y 10.63 ms en todas las configuraciones
probadas. Este valor despreciable frente al tiempo total de ejecucion
(>4,600 ms, representando menos del 0.3%) valida empiricamente la
decision de diseno de delegar la agregacion de frecuencias lexicas a
estructuras locales por worker (map[string]int) y reservar el mutex
unicamente para contadores escalares, evitando el cuello de botella de
sincronizacion reportado por Treviso et al. (2023) en pipelines de NLP
con estado global compartido. La mayor contencion (10.63 ms) se observa
con batch=100, donde el numero de adquisiciones de lock es 10x mayor
que con batch=1,000.

## Sobre la configuracion optima recomendada

Configuracion optima del pipeline concurrente: N=8 workers por etapa
(uno por nucleo logico) y batch_size=1,000 documentos. Esta combinacion
ofrece un speedup de 1.85x para 1M registros (throughput de ~205,000
docs/s) manteniendo el consumo de memoria controlado (113 MB) y la
contencion de sincronizacion despreciable (<7 ms). Si la memoria no
es restriccion, batch_size=5,000 alcanza un speedup de 1.93x a costa
de 192 MB de memoria pico.
