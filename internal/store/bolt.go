package store

import (
	"log"

	"github.com/boltdb/bolt"
)

type Bolt struct {
	db *bolt.DB
}

func getBolt(pth string) *bolt.DB {
	db, err := bolt.Open(pth, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("images"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("products"))
		return err
	})

	return db
}

func (b *Bolt) Put(key, val, bucket []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)
		return bu.Put(key, val)
	})
}

func (b *Bolt) Get(key, bucket []byte, f func(val []byte) error) error {
	return b.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)
		v := bu.Get(key)
		if len(v) == 0 {
			return ErrNotFound
		}
		return f(v)
	})
}

func (b *Bolt) GetAll(bucket []byte, f func(val []byte) error) error {
	return b.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)
		c := bu.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := f(v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *Bolt) Delete(key, bucket []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)
		return bu.Delete(key)

	})
}

func (b *Bolt) DeleteAll(bucket []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucket)
	})
}
