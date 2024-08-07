package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	ID           int
	RequestedAt  time.Time
	Username     string
	RequestedURL string
}

type FileInfo struct {
	ID uuid.UUID
	Request
	Size int64
	Hash uint64
	URL  string
}

var ErrFileHashAlreadyExists = errors.New("file hash already exists")
