package dump_util

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/OffbyteSecure/dump_util/internal/drivers"
	"github.com/OffbyteSecure/dump_util/internal/utils"
	"github.com/OffbyteSecure/dump_util/pkg/dump_util/types"
)

// DumpDatabase: Main entrypoint, modular via Dumper and Writer.
func DumpDatabase(dbType, connStr, backupFile string, opts *types.BackupOptions) error {
	if opts == nil {
		opts = &types.BackupOptions{BatchSize: 5000, MaxWorkers: 5}
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	ctx := context.Background()
	dumper, w, closeFn, err := drivers.CreateDumperAndWriter(dbType, connStr, backupFile, opts)
	if err != nil {
		return err
	}
	defer closeFn()

	opts.Logger.Info("Starting dump", "type", dbType)

	tables, err := dumper.ListTables(ctx)
	if err != nil {
		return err
	}

	// Filter excludes
	tables = utils.FilterTables(tables, opts.Exclude)

	sem := make(chan struct{}, opts.MaxWorkers)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, table := range tables {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			schema, err := dumper.DumpSchema(ctx, t)
			if err != nil {
				opts.Logger.Error("Schema dump failed", "table", t, "error", err)
				return
			}
			if err := w.WriteSchema(t, schema); err != nil {
				opts.Logger.Error("Write schema failed", "table", t, "error", err)
				return
			}

			// Stream data batches
			dataChan := make(chan []string, 1)
			go func() {
				defer close(dataChan)
				for {
					batch, err := dumper.DumpData(ctx, t, opts.BatchSize)
					if err == io.EOF {
						return
					}
					if err != nil {
						opts.Logger.Error("Data batch failed", "table", t, "error", err)
						return
					}
					dataChan <- batch
				}
			}()

			for batch := range dataChan {
				mu.Lock()
				if err := w.WriteData(t, batch); err != nil {
					opts.Logger.Error("Write data failed", "table", t, "error", err)
				}
				mu.Unlock()
			}
		}(table)
	}
	wg.Wait()

	return w.Flush()
}
