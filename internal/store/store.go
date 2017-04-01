package store

import (
	"errors"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"

	"github.com/cswank/store/internal/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Query struct {
	Key     []byte
	Buckets [][]byte
	Val     []byte
}

func NewQuery(opts ...func(*Query)) Query {
	r := &Query{}
	for _, o := range opts {
		o(r)
	}
	return *r
}

func Key(k string) func(*Query) {
	return func(r *Query) {
		r.Key = []byte(k)
	}
}

func Buckets(buckets ...string) func(*Query) {
	return func(r *Query) {
		for _, b := range buckets {
			r.Buckets = append(r.Buckets, []byte(b))
		}
	}
}

func Val(v []byte) func(*Query) {
	return func(r *Query) {
		r.Val = v
	}
}

type storer interface {
	Put([]Query) error
	Get([]Query, func(k, v []byte) error) error
	GetAll(Query, func(k, v []byte) error) error
	Delete([]Query) error
	DeleteAll([]byte) error
	AddBucket(Query) error
	RenameBucket(Query, Query) error
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

func SetDB(d storer) func() {
	return func() {
		db = d
	}
}

func GetImage(bucket, title, size string) ([]byte, error) {
	q := []Query{{Key: []byte(size), Buckets: [][]byte{[]byte("images"), []byte(bucket), []byte(title)}}}
	var img []byte
	return img, db.Get(q, func(k, v []byte) error {
		img = v
		return nil
	})
}
