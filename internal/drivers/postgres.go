package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	// Blank import in go.mod
)

// postgresDumper implements Dumper.
type postgresDumper struct {
	db *sql.DB
}

func (p *postgresDumper) ListTables(ctx context.Context) ([]string, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		rows.Scan(&t)
		tables = append(tables, t)
	}
	return tables, nil
}

func (p *postgresDumper) DumpSchema(ctx context.Context, table string) (string, error) {
	query := fmt.Sprintf(`SELECT 'CREATE TABLE %s (' || string_agg(column_def, ',\n  ') || ');' FROM (
        SELECT column_name || ' ' || data_type || 
        CASE WHEN is_nullable = 'NO' THEN ' NOT NULL' END ||
        CASE WHEN column_default IS NOT NULL THEN ' DEFAULT ' || column_default END AS column_def
        FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = $1 
        ORDER BY ordinal_position
    ) t`, table)
	var schema string
	err := p.db.QueryRowContext(ctx, query, table).Scan(&schema)
	return schema, err
}

func (p *postgresDumper) DumpData(ctx context.Context, table string, batchSize int) ([]string, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1", table)
	rows, err := p.db.QueryContext(ctx, query, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var batch []string
	for rows.Next() {
		// Scan and build INSERT values (optimized scanning)
		values := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		rows.Scan(scanArgs...)
		var valStrs []string
		for _, v := range values {
			if v == nil {
				valStrs = append(valStrs, "NULL")
			} else {
				strVal := fmt.Sprintf("%v", v)
				strVal = strings.ReplaceAll(strVal, "'", "''")
				valStrs = append(valStrs, fmt.Sprintf("'%s'", strVal))
			}
		}
		batch = append(batch, "("+strings.Join(valStrs, ", ")+")")
	}
	if len(batch) == 0 {
		return nil, io.EOF
	}
	return batch, nil
}

func (p *postgresDumper) Close() error { return p.db.Close() }
