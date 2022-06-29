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
}
