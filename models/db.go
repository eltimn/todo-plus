package models

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/tursodatabase/go-libsql"
)

// https://www.alexedwards.net/blog/organising-database-access
// TODO: add sql builder - https://github.com/huandu/go-sqlbuilder
var db *sql.DB

func OpenDB(uri string) (*sql.DB, error) {
	var err error

	db, err = sql.Open("libsql", uri)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}

func CloseDB() error {
	return db.Close()
}

type Env struct {
	users UserModel
}

// Same as ExecContext, but checks that only one row was affected
func ExecOneContext(ctx context.Context, query string, args ...any) error {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}

	return nil
}
