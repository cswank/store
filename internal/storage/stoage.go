package storage

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/cswank/store/internal/config"
)

type Row struct {
	Key     []byte
	Buckets [][]byte
	Val     []byte
}

type storer interface {
	Put([]Row) error
	Get([]Row, func(k, v []byte) error) error
	GetAll(Row, func(k, v []byte) error) error
	Delete([]Row) error
	DeleteAll([]byte) error
	AddBucket(Row) error
	RenameBucket(Row, Row) error
	GetBackup(w http.ResponseWriter) error
}

var (
	ErrNotFound = errors.New("not found")
	db          storer
	cfg         config.Config
)

func Init(c config.Config, opts ...func()) {
	cfg = c
	for _, opt := range opts {
		opt()
	}

	if db == nil {
		b := getBolt(filepath.Join(cfg.DataDir, "db"))
		db = &Bolt{db: b}
	}
}
