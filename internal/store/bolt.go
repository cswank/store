package store

import "github.com/boltdb/bolt"

type Bolt struct {
	db *bolt.DB
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
