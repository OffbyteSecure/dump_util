package utils

import (
    "database/sql"
    "time"
)

// ConfigurePool sets pooling for performance (imported by drivers).
func ConfigurePool(db *sql.DB) {
    db.SetMaxOpenConns(50)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(1 * time.Minute)
}

// FilterTables filters out excluded tables.
func FilterTables(tables []string, exclude []string) []string {
    m := make(map[string]bool)
    for _, e := range exclude {
        m[e] = true
    }
    var filtered []string
    for _, t := range tables {
        if !m[t] {
            filtered = append(filtered, t)
        }
    }
    return filtered
}