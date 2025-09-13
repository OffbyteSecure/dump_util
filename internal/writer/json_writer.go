package writer

import (
	"bufio"
	"encoding/json"
	"io" // Added for map keys
	"time"
)

// JSONWriter for MongoDB-style dumps.
type JSONWriter struct {
	enc *json.Encoder
	w   io.Writer // Underlying writer for flushing
}

func NewJSONWriter(w io.Writer) *JSONWriter {
	// Use buffered writer for better performance
	bw := bufio.NewWriter(w)
	enc := json.NewEncoder(bw)
	enc.SetIndent("", "  ")
	return &JSONWriter{enc: enc, w: bw}
}

func (jw *JSONWriter) WriteHeader(dbType string) error {
	return jw.enc.Encode(map[string]string{"type": dbType, "generated": time.Now().Format(time.RFC3339)})
}

func (jw *JSONWriter) WriteSchema(table, schema string) error { return nil } // No schema

func (jw *JSONWriter) WriteData(table string, data []string) error {
	for _, d := range data {
		if err := jw.enc.Encode(map[string]interface{}{"table": table, "data": d}); err != nil {
			return err
		}
	}
	return nil
}

func (jw *JSONWriter) Flush() error {
	// Flush the encoder's underlying buffered writer
	if bw, ok := jw.w.(*bufio.Writer); ok {
		return bw.Flush()
	}
	// If underlying writer supports Flush (e.g., gzip.Writer), call it
	if flusher, ok := jw.w.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	// Otherwise, no-op (Encoder flushes on Encode)
	return nil
}
