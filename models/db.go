package models

import (
	"database/sql"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
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
