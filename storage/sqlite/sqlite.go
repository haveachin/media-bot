package sqlite

import "database/sql"

type SQLite struct {
	db *sql.DB
}

func New(dsn string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	return &SQLite{
		db: db,
	}, nil
}
