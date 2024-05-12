package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"media-bot/types"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

type SQLite struct {
	db *sql.DB
}

func New(dsn string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := initDB(db); err != nil {
		return nil, err
	}

	return &SQLite{
		db: db,
	}, nil
}

func initDB(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Handle Migrations

	return nil
}

func (db *SQLite) InsertFileInfoRequest(
	ctx context.Context,
	username string,
	requestedURL string,
	fileSize int64,
	fileHash []byte,
) (uuid.UUID, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	var requestID int64
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO requests(username, requested_url) VALUES(?, ?) RETURNING id;`,
		username,
		requestedURL,
	).Scan(&requestID); err != nil {
		return uuid.Nil, err
	}

	fileInfoID := uuid.New()
	if _, err := tx.ExecContext(
		ctx,
		"INSERT INTO file_infos(id, request_id, size, hash) VALUES(?, ?, ?, ?);",
		fileInfoID[:],
		requestID,
		fileSize,
		fileHash[:],
	); err != nil {
		sqlErr, ok := err.(sqlite3.Error)
		if ok && sqlErr.Code == sqlite3.ErrConstraint {
			return uuid.Nil, types.ErrFileHashAlreadyExists
		}

		return uuid.Nil, err
	}

	return fileInfoID, nil
}

func (db *SQLite) FileURLByHash(
	ctx context.Context,
	hash []byte,
) (string, error) {
	var url string
	if err := db.db.QueryRowContext(
		ctx,
		"SELECT url FROM file_infos WHERE hash = ?;",
		hash[:],
	).Scan(&url); err != nil {
		return "", err
	}

	return url, nil
}

func (db *SQLite) InsertFileInfoURL(
	ctx context.Context,
	fileInfoID uuid.UUID,
	url string,
) error {
	if _, err := db.db.ExecContext(ctx,
		`UPDATE file_infos SET url = ? WHERE id = ?;`,
		url,
		fileInfoID[:],
	); err != nil {
		return err
	}

	return nil
}
