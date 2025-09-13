# dump_util

A modular, high-performance Go library for dumping databases (PostgreSQL, MySQL, MongoDB) to files, with CLI support.

## Library Usage
```go
import "github.com/OffbyteSecure/dump_util"

opts := &dump_util.BackupOptions{ /* ... */ }
if err := dump_util.DumpDatabase("postgres", "connstr", "backup.sql", opts); err != nil {
    log.Fatal(err)
}
