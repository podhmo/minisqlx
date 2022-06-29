package minisqlx

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "modernc.org/sqlite"
)

// func Must[T any](x T, err error) T {
// 	if err != nil {
// 		panic(err)
// 	}
// 	return x
// }

type user struct {
	Name string
	Age  int
}

func newDB(ctx context.Context, t *testing.T) (DB, func()) {
	t.Helper()
	db, err := ConnectContext(ctx, "sqlite", ":memory:")
	if err != nil {
		t.Fatalf("connect: %+v", err)
	}

	stmt := `CREATE TABLE users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			age Integer NOT NULL DEFAULT 0
		  );`
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		t.Fatalf("create table: %+v", err)
	}

	return db, func() {
		stmt := `DROP TABLE users;`
		_, err := db.ExecContext(ctx, stmt)
		if err != nil {
			t.Fatalf("drop table: %+v", err)
		}
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	_, teardown := newDB(ctx, t)
	defer teardown()
}

func TestInsertOne(t *testing.T) {
	ctx := context.Background()
	db, teardown := newDB(ctx, t)
	defer teardown()

	var want struct {
		user         user
		lastInsertID int
	}
	want.user.Name = "foo"
	want.user.Age = 20
	want.lastInsertID = 1

	stmt := `INSERT INTO users(name, age) VALUES(?, ?)`
	info, err := db.ExecContext(ctx, stmt, want.user.Name, want.user.Age)
	if err != nil {
		t.Fatalf("unexpected error is found: %+v", err)
	}

	got, err := info.LastInsertId()
	if err != nil {
		t.Errorf("insert data, last inserted Id: %+v", err)
	}
	if want, got := want.lastInsertID, int(got); want != got {
		t.Errorf("last inserted id, want=%d, but got=%d", want, got)
	}
}

// TODO: bulk insert

func TestSelectOne(t *testing.T) {
	ctx := context.Background()
	db, teardown := newDB(ctx, t)
	defer teardown()

	want := user{Name: "foo", Age: 20}
	stmt := `SELECT name, age FROM users WHERE Name=?`

	t.Run("before insert", func(t *testing.T) {
		rows, err := db.QueryxContext(ctx, stmt, want.Name)
		if err != nil {
			t.Fatalf("select one: %+v", err)
		}

		i := 0
		for ; rows.Next(); i++ {
			var u user
			if err := rows.Scan(&u.Name, &u.Age); err != nil { // use StructScan()?
				t.Errorf("rows[%d].Scan(): %+v", err, i)
			}
			if diff := cmp.Diff(u, want); diff != "" {
				t.Errorf("rows[%d] data mismatch (-got +want):\n%s", i, diff)
			}
		}
		if want, got := 0, i; want != got {
			t.Errorf("num of rows want=%v, but got=%v", want, got)
		}
	})

	{
		stmt := `INSERT INTO users(name, age) VALUES(?, ?)`
		if _, err := db.ExecContext(ctx, stmt, want.Name, want.Age); err != nil {
			t.Errorf("insert data, unexpeted error is found: %+v", err)
		}
	}

	t.Run("after insert", func(t *testing.T) {
		rows, err := db.QueryxContext(ctx, stmt, want.Name)
		if err != nil {
			t.Fatalf("select one: %+v", err)
		}

		i := 0
		for ; rows.Next(); i++ {
			var u user
			if err := rows.Scan(&u.Name, &u.Age); err != nil { // use StructScan()?
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
}

func TestGet(t *testing.T) { // Get is shorthand of select one
	ctx := context.Background()
	db, teardown := newDB(ctx, t)
	defer teardown()

	want := user{Name: "foo", Age: 20}
	t.Run("before insert", func(t *testing.T) {
		stmt := `SELECT name, age FROM users WHERE Name=?`
		_, err := GetContext[user](ctx, db, stmt, want.Name)
		if err == nil {
			t.Fatalf("error is expetecd, but nil")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("unexpected error is found: %+v", err)
		}
	})

	{
		stmt := `INSERT INTO users(name, age) VALUES(?, ?)`
		if _, err := db.ExecContext(ctx, stmt, want.Name, want.Age); err != nil {
			t.Errorf("insert data, unexpeted error is found: %+v", err)
		}
	}

	t.Run("after insert", func(t *testing.T) {
		stmt := `SELECT name, age FROM users WHERE name=? order by name desc`
		got, err := GetContext[user](ctx, db, stmt, want.Name)
		if err != nil {
			t.Errorf("unexpected error is found: %+v", err)
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("data mismatch (-got +want):\n%s", diff)
		}
	})
}

func TestGetMany(t *testing.T) { // Get is shorthand of select one
	ctx := context.Background()
	db, teardown := newDB(ctx, t)
	defer teardown()

	age := 20
	want := []user{
		{Name: "foo", Age: age},
		{Name: "bar", Age: age},
	}

	t.Run("before insert", func(t *testing.T) {
		stmt := `SELECT name, age FROM users WHERE age=? ORDER BY name desc`
		got, err := GetManyContext[user](ctx, db, stmt, age)
		if err != nil {
			t.Fatalf("unexpected error is found: %+v", err)
		}
		if want := 0; len(got) != want {
			t.Errorf("want len() == %d, but got %d", want, len(got))
		}
	})

	{
		stmt := `INSERT INTO users(name, age) VALUES (?, ?),(?, ?)`
		if _, err := db.ExecContext(ctx, stmt,
			want[0].Name, want[0].Age,
			want[1].Name, want[1].Age,
		); err != nil {
			t.Errorf("insert data, unexpeted error is found: %+v", err)
		}
	}

	t.Run("after insert", func(t *testing.T) {
		stmt := `SELECT name, age FROM users WHERE age=? ORDER BY name desc`
		got, err := GetManyContext[user](ctx, db, stmt, age)
		if err != nil {
			t.Fatalf("unexpected error is found: %+v", err)
		}
		if want := 2; len(got) != want {
			t.Errorf("want len() == %d, but got %d", want, len(got))
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("data mismatch (-got +want):\n%s", diff)
		}
	})
}
