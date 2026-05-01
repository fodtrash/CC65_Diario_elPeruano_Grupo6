// corpus.go — Tipos de datos y carga del corpus
// Equivale al campo ORIGEN y a la etapa de lectura del modelo Promela.
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ── Constantes del corpus ────────────────────────────────────────────────────

const (
	OrigenReal      = 0 // Registro proveniente del corpus limpio (Diario El Peruano)
	OrigenSintetico = 1 // Registro generado sintéticamente para aumentación
)

// ── Tipos del pipeline ───────────────────────────────────────────────────────

// Document representa un registro del corpus pre-aumentado.
// El campo Origen ya viene asignado desde el corpus fusionado (columna ORIGEN).
type Document struct {
	ID      int
	Sumilla string
	Origen  int // OrigenReal | OrigenSintetico
}

// TokenizedDoc es la salida de la etapa de tokenización BPE.
type TokenizedDoc struct {
	Document
	Tokens []string
}

// LematizedDoc es la salida de la etapa de lematización y normalización.
type LematizedDoc struct {
	TokenizedDoc
	Lemmas []string
}

// ── Carga del corpus ─────────────────────────────────────────────────────────

// LoadCorpus intenta leer el CSV en csvPath.
// Espera columnas: ID (o índice implícito), SUMILLA, ORIGEN.
// Si el archivo no existe o no es legible, genera un corpus sintético
// de nReal registros REAL y nSint registros SINTETICO.
func LoadCorpus(csvPath string, nReal, nSint int) ([]Document, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		// Archivo no encontrado → corpus generado (útil para benchmarks)
		fmt.Printf("[corpus] %s no encontrado — generando corpus (%d REAL + %d SINT)\n",
			csvPath, nReal, nSint)
		return generateCorpus(nReal, nSint), nil
	}
	defer f.Close()

	docs, err := parseCSV(f, nReal+nSint)
	if err != nil {
		return nil, fmt.Errorf("error leyendo CSV: %w", err)
	}
	fmt.Printf("[corpus] Cargados %d registros desde %s\n", len(docs), csvPath)
	return docs, nil
}

// parseCSV lee el CSV y mapea columnas a Document.
// Soporta encabezados en minúsculas o mayúsculas.
// Límite opcional: si limit > 0 se detiene al alcanzarlo.
func parseCSV(r io.Reader, limit int) ([]Document, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Mapeo flexible de columnas (case-insensitive)
	colIdx := map[string]int{}
	for i, h := range headers {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	idCol, hasID := colIdx["id"]
	sumillaCol, hasSumilla := colIdx["sumilla"]
	origenCol, hasOrigen := colIdx["origen"]

	if !hasSumilla {
		return nil, fmt.Errorf("columna SUMILLA no encontrada en CSV")
	}

	var docs []Document
	rowNum := 0

	for {
		if limit > 0 && len(docs) >= limit {
			break
		}
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Fila malformada → se omite con advertencia
			fmt.Printf("[corpus] fila %d omitida: %v\n", rowNum, err)
			rowNum++
			continue
		}

		doc := Document{}

		if hasID {
			doc.ID, _ = strconv.Atoi(strings.TrimSpace(record[idCol]))
		} else {
			doc.ID = rowNum + 1
		}

		doc.Sumilla = strings.TrimSpace(record[sumillaCol])

		if hasOrigen {
			v, _ := strconv.Atoi(strings.TrimSpace(record[origenCol]))
			doc.Origen = v
		} else {
			// Sin columna ORIGEN: asignación por posición (igual que Promela)
			if rowNum < len(docs) { // placeholder; se usa rowNum
				doc.Origen = OrigenReal
			}
		}

		docs = append(docs, doc)
		rowNum++
	}

	return docs, nil
}

// generateCorpus crea un corpus de prueba en memoria.
// Los primeros nReal registros llevan Origen=OrigenReal,
// los siguientes nSint llevan Origen=OrigenSintetico.
// Esto replica exactamente la asignación del modelo Promela:
//
//	doc <= N_REAL → REAL
//	doc >  N_REAL → SINTETICO
func generateCorpus(nReal, nSint int) []Document {
	// Fragmentos representativos del Diario Oficial El Peruano
	realPhrases := []string{
		"Resolución que aprueba la transferencia de partidas del presupuesto institucional",
		"Decreto supremo que modifica el reglamento de la ley de contrataciones del Estado",
		"Ordenanza municipal que regula el uso de espacios públicos en el distrito",
		"Acuerdo de consejo que declara en emergencia el sistema de agua potable",
		"Resolución ministerial que designa al titular de la oficina de administración",
		"Ley que fortalece la transparencia en la gestión de recursos públicos",
	}
	sintPhrases := []string{
		"Norma que establece procedimientos para la adquisición de bienes y servicios",
		"Disposición que regula las condiciones de trabajo en el sector público",
		"Resolución que aprueba el plan operativo institucional para el ejercicio fiscal",
		"Acuerdo que modifica las bases del proceso de selección convocado",
		"Decreto que aprueba el presupuesto analítico del personal de la entidad",
		"Directiva que establece los lineamientos para la gestión de archivos documentales",
	}

	total := nReal + nSint
	docs := make([]Document, 0, total)

	for i := 1; i <= total; i++ {
		origen := OrigenReal
		phrase := realPhrases[(i-1)%len(realPhrases)]
		if i > nReal {
			origen = OrigenSintetico
			phrase = sintPhrases[(i-nReal-1)%len(sintPhrases)]
		}
		docs = append(docs, Document{
			ID:      i,
			Sumilla: fmt.Sprintf("[DOC-%05d] %s.", i, phrase),
			Origen:  origen,
		})
	}
	return docs
}
