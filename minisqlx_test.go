package minisqlx

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"
)

func Must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

func TestIt(t *testing.T) {
	ctx := context.Background()
	db := Must(ConnectContext(ctx, "sqlite", ":memory:"))

	t.Run("create table", func(t *testing.T) {
		stmt := `CREATE TABLE users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			age Integer NOT NULL DEFAULT 0
		  );`

		_, err := db.ExecContext(ctx, stmt)
		if err != nil {
			t.Fatalf("create table: %+v", err)
		}
	})
}
