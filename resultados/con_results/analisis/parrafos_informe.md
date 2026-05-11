# Parrafos pre-redactados para el informe del PC2

Estos parrafos estan escritos en un registro tecnico-academico para el informe.
Todos los datos provienen de los 35 JSON en `resultados/con_results/raw/`,
generados con la corrida mas reciente sobre `dataset_final_1M.csv`.

---

## Sobre el speedup y la metodologia estadistica

Para controlar la variabilidad experimental, cada configuracion se ejecuto cinco veces y se uso media recortada (sin minimo ni maximo) como estimador robusto. El baseline concurrente `N=1, batch=1000` obtuvo **5,434 ms** (CV 0.7%), mientras que el mejor resultado para la familia `batch=1000` fue **N=4** con **2,377 ms** (speedup **2.29x**, eficiencia 57.2%). Con mayor paralelismo se observa rendimiento decreciente: `N=8` marca 2,494 ms (2.18x; 27.2%) y `N=16` sube a 3,012 ms (1.80x; 11.3%). Esta curva confirma que aumentar workers no garantiza mejoras lineales cuando la carga util paralelizable ya fue reducida por memoizacion y la etapa mas lenta del pipeline pasa a dominar el wall-clock.

## Sobre el comportamiento por etapas del pipeline

Las medias recortadas por etapa muestran solapamiento fuerte entre lectura, tokenizacion y lematizacion. Por ejemplo, en `N=4, b=1000` se observa `elapsed_read_ms=2,371`, `elapsed_token_ms=2,374` y `elapsed_lemma_ms=2,365`, con total de 2,377 ms. El mismo patron se repite en las demas configuraciones, lo que indica que la arquitectura en pipeline aprovecha bien la concurrencia entre etapas. En consecuencia, la mejora marginal adicional depende menos de acelerar una etapa aislada y mas de reducir el tramo critico global que queda en cada corrida.

## Sobre la degradacion relativa en N alto

El caso `N=16, b=1000` exhibe la mayor dispersion (CV 10.5%) debido a un outlier de 3,798 ms dentro de sus cinco mediciones. Este comportamiento sugiere sensibilidad a condiciones de ejecucion del entorno (planificador, presion de memoria o ruido del sistema) cuando se incrementa la sobre-suscripcion de goroutines por etapa. En terminos practicos, `N=16` no solo es mas lento que `N=4`, sino tambien menos estable, por lo que no resulta una opcion recomendable para ejecuciones repetibles de benchmarking.

## Sobre el trade-off tiempo/memoria por batch size

Con `N=8` fijo, el mejor tiempo recortado aparece en `batch=5000` con **2,375 ms**, superando a `batch=1000` (2,494 ms) y `batch=100` (2,436 ms). Sin embargo, este beneficio temporal viene acompanado de una penalizacion de memoria: la memoria pico promedio sube a **321 MB** en `batch=5000`, frente a ~258-259 MB en los otros dos lotes. Por tanto, la eleccion de batch debe responder al objetivo operativo: `batch=5000` si se prioriza tiempo absoluto, o `batch=1000` si se busca un balance mas conservador de rendimiento y uso de RAM.

## Sobre sincronizacion y contencion

La metrica `mutex_contention_ms` reporta maximo **0.00 ms** en todas las configuraciones evaluadas, lo que indica ausencia de contencion medible sobre el lock global de contadores en esta tanda. Este resultado respalda el diseno de agregacion local por worker y sugiere que, en el estado actual, el mutex no constituye cuello de botella observable. En otras palabras, las diferencias de rendimiento entre configuraciones se explican principalmente por distribucion de carga, costos de scheduling y politicas de batch, no por bloqueo activo en la seccion critica.

## Sobre determinismo funcional

Las 35 corridas conservaron exactamente los mismos resultados semanticos: `total_tokens=8,847,717`, `total_lemmas_unique=38,201`, `docs_reales=244,779` y `docs_sinteticos=755,221`. Esta invariancia frente a cambios de N y batch confirma que el pipeline concurrente mantiene consistencia funcional y que la paralelizacion afecta el tiempo de ejecucion, no la salida del procesamiento NLP.

## Sobre la comparacion con la linea base secuencial

Tomando como referencia la media recortada secuencial reportada de **613,593 ms**, la mejor configuracion concurrente en `batch=1000` (`N=4`) alcanza una relacion `T_seq/T_conc` de **258.1x**. Este cociente es util como indicador global de mejora empirica, pero debe interpretarse con cautela metodologica: incluye tanto el efecto de paralelizacion como las optimizaciones internas presentes en el flujo concurrente. Para una comparacion estrictamente iso-funcional, ambas implementaciones deberian compartir exactamente la misma estrategia de optimizacion interna.
