# dump_util

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/OffbyteSecure/dump_util)](https://goreportcard.com/report/github.com/OffbyteSecure/dump_util)
[![Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen.svg)](https://pkg.go.dev/github.com/OffbyteSecure/dump_util)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/OffbyteSecure/dump_util)](https://pkg.go.dev/github.com/OffbyteSecure/dump_util)

A modular, high-performance Go library for dumping PostgreSQL, MySQL, and MongoDB databases to SQL or JSON files. Supports schema exports, batched data streaming, compression, concurrency, and table exclusions for efficient backups. Includes a CLI tool for quick dumps without writing code.

## Features

- **Multi-Database Support**: Seamless backups from PostgreSQL, MySQL, and MongoDB using native Go drivers.
- **Schema & Data Dumps**: Generates portable SQL (CREATE TABLE + INSERTs) for relational DBs; JSON arrays for MongoDB collections.
- **Performance Optimized**: Concurrent table dumps (configurable workers), batched exports (up to 10k rows/batch), connection pooling, and streaming to avoid memory bloat.
- **Flexible Options**: Gzip compression, exclude tables/collections, custom batch sizes, and structured logging.
- **CLI Included**: `dumper` binary for command-line backups (e.g., `./dumper postgres --conn "user:pass@host/db" --output backup.sql.gz`).
- **Pure Go**: No external dependencies beyond standard drivers; easy to embed in apps.

Why choose dump_util? Unlike `pg_dump` or `mongodump` (which require binaries), this is a lightweight, Go-native solution for cross-DB backups in microservices or scripts. It's faster than naive `database/sql` loops for large datasets.

## Quick Start

### Installation

```bash
# As a library
go get github.com/OffbyteSecure/dump_util@v0.1.0

# As a CLI tool
go install github.com/OffbyteSecure/dump_util/cmd/dumper@latest

Library UsageImport the package and use DumpDatabase for a simple backup:go

package main

import (
    "log"
    "log/slog"
    "os"

    "github.com/OffbyteSecure/dump_util"
)

func main() {
    opts := &dump_util.BackupOptions{
        Compress:   true,     // Enable gzip compression
        BatchSize:  5000,     // Process 5k rows/documents per batch
        MaxWorkers: 5,        // Concurrent table dumps
        Exclude:    []string{"temp_table", "logs"}, // Skip these tables/collections
        Logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)), // Optional logging
    }

    if err := dump_util.DumpDatabase(
        "postgres",                                   // DB type
        "postgres://user:pass@localhost:5432/mydb?sslmode=disable", // Connection string
        "backup.sql.gz",                              // Output file
        opts,
    ); err != nil {
        log.Fatal(err)
    }
}

Expected Output: A gzipped backup.sql with headers, CREATE TABLE statements, and batched INSERTs (e.g., INSERT INTO users VALUES (...), (...);).For MySQL: Use "mysql" and a DSN like user:pass@tcp(localhost:3306)/mydb.
For MongoDB: Use "mongodb" and a URI like mongodb://localhost:27017, outputting JSON like {"db.users": [{"_id": "...", "name": "Alice"}]}.See pkg.go.dev/github.com/OffbyteSecure/dump_util for full API docs and godoc examples.CLI UsageThe dumper tool wraps the library for shell scripting:bash

# Basic PostgreSQL dump (compressed)
dumper --type postgres \\
       --conn "postgres://user:pass@localhost:5432/mydb?sslmode=disable" \\
       --output backup.sql.gz \\
       --compress \\
       --batch-size 10000 \\
       --workers 8 \\
       --exclude "temp_table,audit_logs"

# MySQL example
dumper --type mysql --conn "user:pass@tcp(localhost:3306)/mydb" --output mysql_backup.sql

# MongoDB (JSON output)
dumper --type mongodb --conn "mongodb://localhost:27017" --output mongo_backup.json.gz

Run dumper --help for all flags. Logs progress to stdout (e.g., "Dumping table 'users': 10k rows processed").Project StructureThis follows the Go Project Layout:/cmd/dumper: CLI entrypoint (Cobra-based).
/internal/drivers: DB-specific implementations (postgres.go, mysql.go, mongodb.go).
/internal/writer: Output formatters (SQLWriter, JSONWriter).
/internal/utils: Helpers (pooling, DSN parsing).
/pkg/dump_util: Public API (BackupOptions, DumpDatabase, interfaces).

ContributingWe welcome bug reports, features, and docs improvements! See CONTRIBUTING.md for details. Run tests with go test ./... and build with go build ./....LicenseMIT License (LICENSE) © 2025 OffbyteSecure. See LICENSE for details.Built with  for efficient DB backups. Questions? Open an issue!
