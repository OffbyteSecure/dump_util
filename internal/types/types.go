package types

import (
	"context"
	"database/sql"
	"log/slog"
)

// Writer defines an extensible output writer (e.g., SQL, JSON).
type Writer interface {
	WriteHeader(dbType string) error
	WriteSchema(table, schema string) error
	WriteData(table string, data []string) error // Batched data lines
	Flush() error
}

// Dumper defines the interface for database-specific dumping.
type Dumper interface {
	ListTables(ctx context.Context) ([]string, error)
	DumpSchema(ctx context.Context, table string) (string, error)                // CREATE TABLE string
	DumpData(ctx context.Context, table string, batchSize int) ([]string, error) // Batched INSERT/JSON strings
	Close() error
}

// BackupOptions with performance tunables.
type BackupOptions struct {
	Compress   bool
	BatchSize  int // Default: 5000
	MaxWorkers int // Concurrent tables (default: 5)
	Exclude    []string
	Logger     *slog.Logger
	SQLPool    sql.DBStats
}
