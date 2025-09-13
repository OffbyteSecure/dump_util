package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoDumper implements Dumper.
type mongoDumper struct {
	client *mongo.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func (m *mongoDumper) ListTables(ctx context.Context) ([]string, error) {
	// Flatten all collections across DBs (customize as needed)
	dbs, err := m.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var allTables []string
	for _, dbName := range dbs {
		if strings.HasPrefix(dbName, "system.") {
			continue
		}
		cols, err := m.client.Database(dbName).ListCollectionNames(ctx, bson.D{})
		if err != nil {
			continue
		}
		for _, col := range cols {
			allTables = append(allTables, dbName+"."+col)
		}
	}
	return allTables, nil
}

func (m *mongoDumper) DumpSchema(ctx context.Context, table string) (string, error) {
	return "", nil // No schema for MongoDB
}

func (m *mongoDumper) DumpData(ctx context.Context, table string, batchSize int) ([]string, error) {
	parts := strings.SplitN(table, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid table: %s", table)
	}
	dbName, collName := parts[0], parts[1]
	opts := options.Find().SetBatchSize(int32(batchSize)).SetMaxTime(30 * time.Second)
	cursor, err := m.client.Database(dbName).Collection(collName).Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batch []string
	for cursor.Next(ctx) {
		var doc bson.Raw
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		batch = append(batch, string(jsonBytes))
		if len(batch) >= batchSize {
			return batch, nil
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if len(batch) == 0 {
		return nil, io.EOF
	}
	return batch, nil
}

func (m *mongoDumper) Close() error {
	m.cancel()
	return m.client.Disconnect(m.ctx)
}
