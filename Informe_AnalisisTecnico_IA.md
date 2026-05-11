# Informe de Análisis Técnico — CC65 - PC2: Pipeline NLP del Diario Oficial El Peruano

## Metadatos del Análisis
| Campo              | Valor                          |
|--------------------|--------------------------------|
| Repositorio        | CC65_Diario_elPeruano_Grupo6   |
| Fecha de análisis  | 10 de Mayo de 2026
| Versión de Go      | 1.20                           |
| Analista           | Claude — Ingeniero Senior Go   |

---

## Resumen Ejecutivo
El proyecto presenta una implementación muy sólida y bien estructurada de un pipeline concurrente (Fan-Out/Fan-In) para procesamiento de lenguaje natural (NLP). Destaca positivamente la segregación del dominio en el paquete `internal/nlp`, el uso de cachés locales por worker para evitar la contención y una excelente metodología de benchmarking. La madurez técnica general es alta (8/10). No obstante, existen oportunidades de mejora relacionadas con el uso intensivo de memoria en rutas críticas (allocations en el cálculo de n-gramas), la falta de propagación de contextos (`context.Context`) para cancelación segura, y el uso de `sync.Mutex` en lugar de operaciones atómicas para contadores, aunque esto último esté justificado por su mapeo al modelo académico Promela.

## Dashboard de Hallazgos

| Dimensión          | Crítico | Alto | Medio | Bajo | Info | Total |
|--------------------|---------|------|-------|------|------|-------|
| Calidad de Código  | 0       | 0    | 0     | 1    | 0    | 1     |
| Concurrencia       | 0       | 0    | 1     | 1    | 0    | 2     |
| Seguridad          | 0       | 0    | 1     | 0    | 0    | 1     |
| Arquitectura       | 0       | 0    | 0     | 0    | 1    | 1     |
| Rendimiento        | 0       | 1    | 0     | 0    | 0    | 1     |
| Deuda Técnica      | 0       | 0    | 1     | 0    | 0    | 1     |
| **TOTAL**          | **0**   | **1**| **3** | **2**| **1**| **7** |

---

## 1. Calidad de Código

### GAP-CC-001: Ausencia de sanitización de rutas en flags CLI
- **Severidad:** BAJO
- **Archivo(s):** `src/concurrent/main.go` (líneas 61-66)
- **Descripción:** Las banderas (flags) de línea de comandos utilizan rutas por defecto relativas al directorio de ejecución que no pasan por `filepath.Clean`. Esto puede generar errores cruzados entre sistemas operativos (Windows vs Linux) y facilita la inyección de rutas inválidas.
- **Evidencia:**
  ```go
  inputFile := flag.String("input", "data/dataset_final_1M.csv", "Ruta al CSV de entrada")
  outputFile := flag.String("output", "resultados/con_results/concurrent_metrics.json", "Ruta del JSON de métricas")
  ```
- **Impacto:** Posibles fallos al ejecutar la herramienta desde directorios distintos a `src/`, provocando que el programa lance un panic y aborte de inmediato.
- **Recomendación:**
  ```go
  import "path/filepath"
  // ...
  cleanInput := filepath.Clean(*inputFile)
  f, err := os.Open(cleanInput)
  ```
- **Esfuerzo:** CORTO PLAZO
- **Referencia:** Go Code Review Comments / OWASP Path Traversal.

---

## 2. Patrones de Concurrencia

### GAP-PC-001: Falta de control de ciclo de vida con context.Context
- **Severidad:** MEDIO
- **Archivo(s):** `src/concurrent/main.go` (Pipeline general)
- **Descripción:** El pipeline confía ciegamente en que la lectura del CSV terminará correctamente y cerrará los canales en cascada. Si el Lector experimenta un error irrecuperable a la mitad del archivo (distinto a EOF) y el programa no hiciera un `log.Fatalf`, los workers se quedarían bloqueados (goroutine leak) esperando datos que nunca llegarán, al no existir un mecanismo de propagación de cancelación.
- **Evidencia:**
  ```go
  // No hay context.Context pasado a los workers
  for batch := range chTokens { ... }
  ```
- **Impacto:** Imposibilidad de abortar el procesamiento de manera controlada (graceful shutdown) frente a señales del sistema operativo (SIGINT/SIGTERM) o errores de E/S.
- **Recomendación:**
  ```go
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  // En el select del worker:
  select {
  case <-ctx.Done():
      return
  case batch, ok := <-chTokens:
      if !ok { return }
      // procesar batch
  }
  ```
- **Esfuerzo:** CORTO PLAZO
- **Referencia:** Go Concurrency Patterns: Pipelines and cancellation

### GAP-PC-002: Sincronización sobre-restrictiva (Mutex vs Atomics)
- **Severidad:** BAJO
- **Archivo(s):** `src/concurrent/main.go` (líneas 49-55, 140-142)
- **Descripción:** Se utiliza `sync.Mutex` para proteger incrementos numéricos en `GlobalCounters`. Aunque en la documentación se justifica explícitamente para mapear la semántica de Promela (`chan mutex = [1] of { bit }`), en un desarrollo idiomático en Go, los contadores globales simples deben gestionarse mediante `sync/atomic`.
- **Evidencia:**
  ```go
  gc.mu.Lock()
  gc.tokensGlobales += batchTokens
  gc.mu.Unlock()
  ```
- **Impacto:** Ligero incremento en la latencia de contención (medido por el equipo en ~6.4ms para N=8). Aunque es despreciable en este contexto, es un antipatrón arquitectónico en Go puro.
- **Recomendación:**
  ```go
  atomic.AddInt64(&gc.tokensGlobales, batchTokens)
  ```
- **Esfuerzo:** INMEDIATO (Se puede mantener si se prioriza 100% la fidelidad al modelo formal, pero se documenta como deuda técnica del lenguaje).
- **Referencia:** Effective Go / `sync/atomic` package.

---

## 3. Seguridad

### GAP-SEC-001: Ausencia de validación al abrir/escribir archivos externos
- **Severidad:** MEDIO
- **Archivo(s):** `src/concurrent/main.go` (línea 87 y 268)
- **Descripción:** El pipeline abre archivos de entrada y salida basados puramente en la entrada del usuario (`*inputFile`, `*outputFile`) sin asegurar que las rutas se mantengan dentro de los límites del directorio del proyecto. 
- **Evidencia:**
  ```go
  f, err := os.Open(*inputFile) // Podría ser ../../../../../etc/passwd
  // ...
  if err := os.WriteFile(*outputFile, outJSON, 0644); err != nil {
  ```
- **Impacto:** Si la aplicación se expusiera como servicio o el entorno de ejecución fuera restrictivo, permitiría ataques de Path Traversal o sobreescritura de archivos arbitrarios (ej. reescribir ejecutables del sistema).
- **Recomendación:**
  Asegurar que la ruta base esté contenida dentro de una carpeta permitida utilizando `filepath.Abs` y verificando el prefijo.
- **Esfuerzo:** CORTO PLAZO
- **Referencia:** CWE-22: Improper Limitation of a Pathname to a Restricted Directory ('Path Traversal').

---

## 4. Arquitectura y Patrones Distribuidos

### GAP-ARQ-001: Carencia de Logging Estructurado
- **Severidad:** INFORMATIVO
- **Archivo(s):** `src/concurrent/main.go` (varios)
- **Descripción:** Se utiliza `log.Printf` para la salida de la aplicación. En el desarrollo moderno de backend, especialmente en pipelines de procesamiento de datos orientados a observabilidad, se prefiere un log estructurado (JSON).
- **Evidencia:**
  ```go
  log.Printf("Pipeline concurrente: input=%s workers(T=%d,L=%d) batch=%d buffer=%d", ...)
  ```
- **Impacto:** Dificulta la ingesta automatizada de logs en sistemas como ELK (Elasticsearch, Logstash, Kibana) o Datadog.
- **Recomendación:**
  Utilizar la librería `golang.org/x/exp/slog` (o `log/slog` si se actualiza a Go 1.21+) o librerías como `uber-go/zap` para emitir logs estructurados.
- **Esfuerzo:** CORTO PLAZO
- **Referencia:** 12-Factor App (Logs as event streams).

---

## 5. Rendimiento y Escalabilidad

### GAP-REND-001: Memory Allocations excesivas en Hot Paths (Lematización)
- **Severidad:** ALTO
- **Archivo(s):** `src/internal/nlp/lemmatize.go` (líneas 37-58)
- **Descripción:** La función `ngramVector` es llamada por cada token no cacheado. Dentro de esta función, se asigna un `map[string]float64` dinámicamente usando `make(map[string]float64, len(runes)*2)`. Los mapas en Go son costosos de instanciar y provocan una alta presión sobre el Garbage Collector (GC) cuando se generan millones de veces (incluso con la excelente caché a nivel de worker).
- **Evidencia:**
  ```go
  func ngramVector(tok string) map[string]float64 {
      // ...
      counts := make(map[string]float64, len(runes)*2) // <--- High allocation rate
  ```
- **Impacto:** Incremento notable del pico de memoria (reportado en ~192MB) y reducción del throughput por paradas del Garbage Collector.
- **Recomendación:**
  Sustituir el uso del `map` por un `[]struct{ ngram string; count float64 }` y ordenarlo, o usar un pool de memoria (`sync.Pool`) para los mapas generados de forma intermedia si el vocabulario crece drásticamente. Dado que los n-gramas son pocos, iterar un slice pre-dimensionado es más amigable para las cachés de CPU (L1/L2) que calcular hashes de strings.
- **Esfuerzo:** MEDIO PLAZO
- **Referencia:** Go Profiling and Optimization / Avoiding Map allocations in hot loops.

---

## 6. Deuda Técnica y Mantenibilidad

### GAP-DT-001: Falta de integración de análisis estático y Makefile
- **Severidad:** MEDIO
- **Archivo(s):** Estructura del repositorio.
- **Descripción:** La ejecución de pruebas, compilación y benchmarking depende de scripts de PowerShell (`run_benchmarks.ps1`) y convenciones documentadas en el `README.md`. No existe un archivo de configuración para `golangci-lint` ni un `Makefile`.
- **Evidencia:** 
  No existen archivos `.golangci.yml` ni `Makefile` en el árbol de directorios `src/`.
- **Impacto:** Pérdida de estandarización en entornos Linux/macOS (PowerShell no es universal en CI/CD). Falta de garantías continuas contra data races o code smells.
- **Recomendación:**
  Crear un `Makefile` con comandos estándar (`make test`, `make bench`, `make lint`) y añadir un pipeline de GitHub Actions o GitLab CI que ejecute `golangci-lint run ./...` y `go test -race ./...` en cada PR.
- **Esfuerzo:** CORTO PLAZO
- **Referencia:** Go community standards for repository structure.

---

## 7. Fortalezas Identificadas

- ✅ **Patrón Pipeline impecable:** Excelente gestión de `sync.WaitGroup` acoplado al cierre de canales en goroutines dedicadas. Evita el bloqueo del sistema (deadlocks) magistralmente.
- ✅ **Optimización de Caché Local:** La decisión en `src/concurrent/main.go` de usar un `lemmaCache := make(map[string]string, 8192)` local por goroutine en la etapa de lematización elimina por completo la necesidad de un Mutex global (que mataría el rendimiento) y reduce drásticamente el uso de CPU, demostrando un profundo entendimiento de cuellos de botella.
- ✅ **Aislamiento de Dominio Pura (Pure Functions):** El paquete `internal/nlp` no tiene estado global ni dependencias de concurrencia. Es 100% thread-safe y testeable, lo cual se evidencia en `nlp_test.go`.
- ✅ **Formalismo Académico y Benchmarking:** La trazabilidad de los JSONs y el análisis profundo del Speedup y Ley de Amdahl denotan un rigor técnico de nivel senior.

---

## 8. Hoja de Ruta de Remediación

### Acciones Inmediatas (sprint actual)
| ID        | Hallazgo            | Responsable sugerido | Estimación |
|-----------|---------------------|----------------------|------------|
| GAP-PC-001| Implementar `context.Context` en pipeline | Ingeniero Go Backend | 4 horas |

### Corto Plazo (próximo mes)
| ID        | Hallazgo            | Responsable sugerido | Estimación |
|-----------|---------------------|----------------------|------------|
| GAP-CC-001| Sanitizar rutas con `filepath.Clean` | Desarrollador | 1 hora |
| GAP-SEC-001| Validación estricta de Path Traversal | Seguridad / Backend  | 2 horas |
| GAP-DT-001| Añadir `Makefile` y `.golangci.yml` | DevSecOps / SRE      | 3 horas |

### Largo Plazo (roadmap técnico)
| ID        | Hallazgo            | Responsable sugerido | Estimación |
|-----------|---------------------|----------------------|------------|
| GAP-REND-001| Refactorizar `ngramVector` (eliminar Maps)| Ing. Optimización Go | 1-2 días |
| GAP-ARQ-001| Migrar a logs estructurados (`slog`) | Ingeniero Go Backend | 1 día |

---

## 9. Métricas de Calidad Estimadas

| Métrica                          | Estado Actual | Objetivo Sugerido |
|----------------------------------|---------------|-------------------|
| Cobertura de pruebas (%)         | ~60-70% (Estimado) | ≥ 80%        |
| Archivos sin manejo de errores   | 1 (os.MkdirAll/os.WriteFile) | 0  |
| Goroutine leaks detectadas       | 0 (Actualmente limpio) | 0        |
| Dependencias con CVEs            | 0 (Librería estándar) | 0         |
| Funciones con complejidad > 15   | 0             | 0                 |
| Paquetes sin documentación       | 0             | 0                 |

---

## 10. Referencias Técnicas
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [OWASP Go Security Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_SCA_Cheat_Sheet.html)
- [golangci-lint](https://golangci-lint.run/)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)