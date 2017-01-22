package store

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/cswank/store/internal/config"
)

var (
	DefaultPrice string
)

type Row struct {
	Key     []byte
	Buckets [][]byte
	Val     []byte
}

func NewRow(opts ...func(*Row)) Row {
	r := &Row{}
	for _, o := range opts {
		o(r)
	}
	return *r
}

func Key(k string) func(*Row) {
	return func(r *Row) {
		r.Key = []byte(k)
	}
}

func Buckets(buckets ...string) func(*Row) {
	return func(r *Row) {
		for _, b := range buckets {
			r.Buckets = append(r.Buckets, []byte(b))
		}
	}
}

func Val(v []byte) func(*Row) {
	return func(r *Row) {
		r.Val = v
	}
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

	DefaultPrice = cfg.DefaultPrice
}

func SetDB(d storer) func() {
	return func() {
		db = d
	}
}

func GetImage(bucket, title, size string) ([]byte, error) {
	q := []Row{{Key: []byte(size), Buckets: [][]byte{[]byte("images"), []byte(bucket), []byte(title)}}}
	var img []byte
	return img, db.Get(q, func(k, v []byte) error {
		img = v
		return nil
	})
}
