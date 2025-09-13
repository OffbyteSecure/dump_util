package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// mysqlDumper implements Dumper.
type mysqlDumper struct {
	db     *sql.DB
	dbName string
}

func (m *mysqlDumper) ListTables(ctx context.Context) ([]string, error) {
	query := `SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'`
	rows, err := m.db.QueryContext(ctx, query, m.dbName)
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

func (m *mysqlDumper) DumpSchema(ctx context.Context, table string) (string, error) {
	var createTable, schema string
	err := m.db.QueryRowContext(ctx, "SHOW CREATE TABLE `%s`.`%s`", m.dbName, table).Scan(&createTable, &schema)
	return schema, err
}

func (m *mysqlDumper) DumpData(ctx context.Context, table string, batchSize int) ([]string, error) {
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT %d", m.dbName, table, batchSize)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var batch []string
	for rows.Next() {
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

func (m *mysqlDumper) Close() error { return m.db.Close() }
