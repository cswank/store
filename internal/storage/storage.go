package storage

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/cswank/store/internal/config"
	"golang.org/x/crypto/bcrypt"
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

type hasher func([]byte, int) ([]byte, error)
type comparer func([]byte, []byte) error

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	db               storer
	cfg              config.Config
	hash             hasher
	compare          comparer
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

	if hash == nil {
		hash = bcrypt.GenerateFromPassword
	}

	if compare == nil {
		compare = bcrypt.CompareHashAndPassword
	}
}

func DB(d storer) func() {
	return func() {
		db = d
	}
}

func Hash(h hasher) func() {
	return func() {
		hash = h
	}
}

func Compare(c comparer) func() {
	return func() {
		compare = c
	}
}
