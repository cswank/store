package store

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/boltdb/bolt"
)

type Config struct {
	Port      int    `env:"STORE_PORT" envDefault:"8080"`
	LogOutput string `env:"STORE_LOGOUTPUT" envDefault:"stdout"`
	DataDir   string `env:"STORE_DATADIR" envDefault:"/var/log/store"`
}

var (
	ErrNotFound = errors.New("not found")
	db          *bolt.DB
)

func Init(cfg Config, opts ...func()) {
	for _, opt := range opts {
		opt()
	}

	if db == nil {
		db = getDB(filepath.Join(cfg.DataDir, "db"))
	}
}

func DB(d *bolt.DB) func() {
	return func() {
		db = d
	}
}

func GetDB() *bolt.DB {
	return db
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
		_, err = tx.CreateBucketIfNotExists([]byte("items"))
		return err
	})

	return db
}
