package store

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/boltdb/bolt"
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
		d := getDB(filepath.Join(cfg.DataDir, "db"))
		db = &Bolt{db: d}
	}
}

func Load() {
	i := Items{}
	if err := i.Load(filepath.Join(cfg.DataDir, "items")); err != nil {
		log.Fatal("store init failed: ", err)
	}
	SetItems(&i)
}

func SetDB(d Storer) func() {
	return func() {
		db = d
	}
}

func getDB(pth string) *bolt.DB {
	db, err := bolt.Open(pth, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("products"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("items"))
		return err
	})

	return db
}
