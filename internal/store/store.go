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

var (
	ErrNotFound = errors.New("not found")
	db          *bolt.DB
	cfg         Config
)

func Init(c Config, opts ...func()) {
	cfg = c
	for _, opt := range opts {
		opt()
	}

	if db == nil {
		db = getDB(filepath.Join(cfg.DataDir, "db"))
	}

	i := Items{}
	if err := i.Load(filepath.Join(cfg.DataDir, "items")); err != nil {
		log.Fatal(err)
	}
	SetItems(&i)
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
