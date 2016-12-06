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

		b, err := tx.CreateBucketIfNotExists([]byte("images"))
		if err != nil {
			return err
		}

		if _, err = b.CreateBucketIfNotExists([]byte("background")); err != nil {
			return err
		}

		if _, err = b.CreateBucketIfNotExists([]byte("products")); err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("products"))
		return err
	})

	return db
}

func (b *Bolt) AddBucket(row Row) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bu, err := b.getOrCreateBucket(tx, row.Buckets)
		if err != nil {
			return err
		}
		_, err = bu.CreateBucketIfNotExists(row.Key)
		return err
	})
}

func (b *Bolt) Put(rows []Row) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		var err error
		for _, r := range rows {
			bu, err := b.getOrCreateBucket(tx, r.Buckets)
			if err != nil {
				return err
			}
			if r.Key != nil && r.Val != nil {
				err = bu.Put(r.Key, r.Val)
			}
		}
		return err
	})
}

func (b *Bolt) getOrCreateBucket(tx *bolt.Tx, buckets [][]byte) (*bolt.Bucket, error) {
	var bu *bolt.Bucket
	for i, n := range buckets {
		var err error
		if i == 0 {
			bu, err = tx.CreateBucketIfNotExists(n)
			if err != nil {
				return nil, err
			}
		} else {
			bu, err = bu.CreateBucketIfNotExists(n)
			if err != nil {
				return nil, err
			}
		}
	}
	return bu, nil
}

func (b *Bolt) getBucket(tx *bolt.Tx, buckets [][]byte) (*bolt.Bucket, error) {
	var bu *bolt.Bucket
	for i, n := range buckets {
		if i == 0 {
			bu = tx.Bucket(n)
		} else {
			bu = bu.Bucket(n)
		}
		if bu == nil {
			return nil, ErrNotFound
		}
	}
	return bu, nil
}

func (b *Bolt) Get(rows []Row, f func(key, val []byte) error) error {
	return b.db.View(func(tx *bolt.Tx) error {
		for _, r := range rows {
			bu, err := b.getBucket(tx, r.Buckets)
			if err != nil {
				return ErrNotFound
			}
			v := bu.Get(r.Key)
			if len(v) == 0 {
				return ErrNotFound
			}
			if err := f(r.Key, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *Bolt) GetAll(r Row, f func(key, val []byte) error) error {
	return b.db.View(func(tx *bolt.Tx) error {
		bu, err := b.getBucket(tx, r.Buckets)
		if err != nil {
			return err
		}
		c := bu.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := f(k, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *Bolt) Delete(rows []Row) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		for _, r := range rows {
			var key []byte
			var f func([]byte) error
			if len(r.Key) > 0 { //delete key/val
				key = r.Key
				bu, err := b.getBucket(tx, r.Buckets)
				if err != nil {
					return err
				}
				f = bu.Delete
			} else { //delete bucket
				l := len(r.Buckets)
				key = r.Buckets[l-1]
				bu, err := b.getBucket(tx, r.Buckets[:l-1])
				if err != nil {
					return err
				}
				f = bu.DeleteBucket
			}

			if err := f(key); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *Bolt) DeleteAll(bucket []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucket)
	})
}
