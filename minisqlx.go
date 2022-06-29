package minisqlx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

var (
	Open = sqlx.Open
)

func ConnectContext(ctx context.Context, driverName, dataSourceName string) (DB, error) {
	return sqlx.ConnectContext(ctx, driverName, dataSourceName)
}

type Tx = sqlx.Tx
type Conn = sqlx.Conn

type NamedStmt = sqlx.NamedStmt
type Stmt = sqlx.Stmt
type Row = sqlx.Row
type Rows = sqlx.Rows

type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func Get[T any](ctx context.Context, db DB, query string, args ...interface{}) (T, error) {
	var dest T
	if err := db.GetContext(ctx, &dest, query, args...); err != nil {
		return dest, err
	}
	return dest, nil
}

func GetMany[T any](ctx context.Context, db DB, query string, args ...interface{}) ([]T, error) {
	var r []T
	rows, err := db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ob T
		if err := rows.StructScan(&ob); err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		r = append(r, ob)
	}
	return r, nil
}
