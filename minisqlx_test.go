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

	t.Run("insert data", func(t *testing.T) {
		var want struct {
			name         string
			age          int
			lastInsertID int
		}
		want.name = "foo"
		want.age = 20
		want.lastInsertID = 1

		stmt := `INSERT INTO users(name, age) VALUES(?, ?)`
		info, err := db.ExecContext(ctx, stmt, want.name, want.age)
		if err != nil {
			t.Fatalf("insert data: %+v", err)
		}
		got, err := info.LastInsertId()
		if err != nil {
			t.Errorf("insert data, last inserted Id: %+v", err)
		}
		if want, got := want.lastInsertID, int(got); want != got {
			t.Errorf("last inserted id, want=%d, but got=%d", want, got)
		}
	})

	// TODO: bulk insert
	t.Run("select one", func(t *testing.T) {
		var want struct {
			name string
			age  int
		}
		want.name = "foo"
		want.age = 20

		stmt := `SELECT name, age FROM users WHERE name=?`

		rows, err := db.QueryxContext(ctx, stmt, want.name)
		if err != nil {
			t.Fatalf("select one: %+v", err)
		}

		var u struct {
			name string
			age  int
		}

		i := 0
		for ; rows.Next(); i++ {
			if err := rows.Scan(&u.name, &u.age); err != nil {
				t.Errorf("rows[%d].Scan(): %+v", err, i)
			}

			if want, got := want.name, u.name; want != got {
				t.Errorf("user.Name want=%v, but got=%v", want, got)
			}
			if want, got := want.age, u.age; want != got {
				t.Errorf("user.age want=%v, but got=%v", want, got)
			}
		}

		if want, got := 1, i; want != got {
			t.Errorf("num of rows want=%v, but got=%v", want, got)
		}
	})
}
