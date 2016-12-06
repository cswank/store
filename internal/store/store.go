package store

import (
	"errors"
	"path/filepath"
)

var (
	DefaultPrice string
)

type Config struct {
	LogOutput    string `env:"STORE_LOGOUTPUT" envDefault:"stdout"`
	DataDir      string `env:"STORE_DATADIR" envDefault:"/var/log/store"`
	DefaultPrice string `env:"STORE_DEFAULT_PRICE" envDefault:"0.00"`
}

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
}

var (
	ErrNotFound = errors.New("not found")
	db          storer
	cfg         Config
)

func Init(c Config, opts ...func()) {
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
