package drivers

import (
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/OffbyteSecure/dump_util/internal/types"
	"github.com/OffbyteSecure/dump_util/internal/utils"
	writerpkg "github.com/OffbyteSecure/dump_util/internal/writer" // Aliased to avoid clash
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// CreateDumperAndWriter initializes based on type.
func CreateDumperAndWriter(dbType, connStr, backupFile string, opts *types.BackupOptions) (types.Dumper, types.Writer, func() error, error) {
	var dumper types.Dumper
	var w types.Writer
	var closeFn func() error = func() error { return nil }

	f, err := os.Create(backupFile)
	if err != nil {
		return nil, nil, nil, err
	}
	var underlying io.Closer = f // Track for close

	if opts.Compress {
		gw := gzip.NewWriter(f)
		underlying = gw // Close gzip before file
		if dbType == "mongodb" {
			w = writerpkg.NewJSONWriter(gw)
		} else {
			w = writerpkg.NewSQLWriter(gw)
		}
		closeFn = func() error {
			if err := underlying.Close(); err != nil {
				return err // Propagate close error
			}
			return f.Close()
		}
	} else {
		if dbType == "mongodb" {
			w = writerpkg.NewJSONWriter(f)
		} else {
			w = writerpkg.NewSQLWriter(f)
		}
		closeFn = f.Close
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	switch dbType {
	case "postgres":
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		utils.ConfigurePool(db)
		if err := db.PingContext(ctx); err != nil {
			db.Close()
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		dumper = &postgresDumper{db: db}
		closeFn = func() error {
			db.Close()
			if err := underlying.Close(); err != nil {
				return err
			}
			if closer, ok := w.(io.Closer); ok {
				return closer.Close()
			}
			return nil
		}
	case "mysql":
		parsed, err := utils.ParseMySQLDSN(connStr)
		if err != nil {
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", parsed["user"], parsed["pass"], parsed["host"], parsed["port"], parsed["db"])
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		utils.ConfigurePool(db)
		if err := db.PingContext(ctx); err != nil {
			db.Close()
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		dumper = &mysqlDumper{db: db, dbName: parsed["db"]}
		closeFn = func() error {
			db.Close()
			if err := underlying.Close(); err != nil {
				return err
			}
			if closer, ok := w.(io.Closer); ok {
				return closer.Close()
			}
			return nil
		}
	case "mongodb":
		client, err := mongo.NewClient(options.Client().ApplyURI(connStr).SetMaxPoolSize(uint64(opts.MaxWorkers * 2)))
		if err != nil {
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		if err := client.Connect(ctx); err != nil {
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			client.Disconnect(ctx)
			if closeErr := closeFn(); closeErr != nil {
				// Optionally log closeErr
			}
			return nil, nil, nil, err
		}
		dumper = &mongoDumper{client: client, ctx: ctx, cancel: cancel}
		closeFn = func() error {
			client.Disconnect(ctx)
			cancel()
			if err := underlying.Close(); err != nil {
				return err
			}
			if closer, ok := w.(io.Closer); ok {
				return closer.Close()
			}
			return nil
		}
	default:
		if closeErr := closeFn(); closeErr != nil {
			// Optionally log closeErr
		}
		return nil, nil, nil, fmt.Errorf("unsupported db_type: %s", dbType)
	}
	cancel() // Cancel after connection

	if err := w.WriteHeader(dbType); err != nil {
		if closeErr := closeFn(); closeErr != nil {
			// Optionally log closeErr
		}
		return nil, nil, nil, err
	}
	return dumper, w, closeFn, nil
}
