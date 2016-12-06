package store

import (
	"errors"
	"path/filepath"
)

type Config struct {
	LogOutput string `env:"STORE_LOGOUTPUT" envDefault:"stdout"`
	DataDir   string `env:"STORE_DATADIR" envDefault:"/var/log/store"`
}

type Storer interface {
	Put([]byte, []byte, []byte) error
	Get([]byte, []byte, func(v []byte) error) error
	GetAll([]byte, func(v []byte) error) error
	Delete([]byte, []byte) error
	DeleteAll([]byte) error
}

var (
	ErrNotFound = errors.New("not found")
	db          Storer
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
}

func SetDB(d Storer) func() {
	return func() {
		db = d
	}
}
