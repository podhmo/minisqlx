package minisqlx

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		type user struct {
			Name string
			Age  int
		}

		want := user{Name: "foo", Age: 20}
		stmt := `SELECT name, age FROM users WHERE Name=?`

		rows, err := db.QueryxContext(ctx, stmt, want.Name)
		if err != nil {
			t.Fatalf("select one: %+v", err)
		}

		i := 0
		for ; rows.Next(); i++ {
			var u user
			if err := rows.Scan(&u.Name, &u.Age); err != nil {
				t.Errorf("rows[%d].Scan(): %+v", err, i)
			}
			if diff := cmp.Diff(u, want); diff != "" {
				t.Errorf("rows[%d] data mismatch (-got +want):\n%s", i, diff)
			}
		}
		if want, got := 1, i; want != got {
			t.Errorf("num of rows want=%v, but got=%v", want, got)
		}
	})
	// TODO: All and One
}
