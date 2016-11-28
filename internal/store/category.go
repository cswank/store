package store

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

type Category struct {
	ID    string `json:"id"`
	OldID string `json:"-"`
}

func GetCategories() (map[string][]string, error) {
	cats := make(map[string][]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		c := b.Cursor()
		for cat, _ := c.First(); cat != nil; cat, _ = c.Next() {
			var subCats []string
			subB := b.Bucket([]byte(cat))
			c := subB.Cursor()
			for subCat, _ := c.First(); subCat != nil; subCat, _ = c.Next() {
				subCats = append(subCats, string(subCat))
			}
			cats[string(cat)] = subCats
		}
		return nil
	})
	return cats, nil
}

func GetSubCatetory(cat, subCat string, page int) ([]Item, error) {
	var items []Item
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		for _, name := range [][]byte{[]byte(cat), []byte(subCat), []byte(fmt.Sprintf("%d", page))} {
			b = b.Bucket(name)
			if b == nil {
				return ErrNotFound
			}
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var i Item
			if err := json.Unmarshal(v, &i); err != nil {
				return err
			}
			items = append(items, i)
		}
		return nil
	})
	fmt.Println("get sub  cat", items)
	return items, err
}

func (c *Category) Save() error {
	return db.Update(func(tx *bolt.Tx) error {
		if c.OldID != "" && c.ID != c.OldID {
			parent := tx.Bucket([]byte("items"))
			return moveBucket(parent, parent, []byte(c.OldID), []byte(c.ID))
		}
		_, err := tx.CreateBucketIfNotExists([]byte(c.ID))
		return err
	})
}

func (c *Category) Delete() error {
	if c.ID == "" {
		return errors.New("can't delete category with no ID")
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.DeleteBucket([]byte(c.ID))
	})
}

// moveBucket moves the inner bucket with key 'oldkey' to a new bucket with key 'newkey'
// must be used within an Update-transaction
func moveBucket(oldParent, newParent *bolt.Bucket, oldkey, newkey []byte) error {
	oldBuck := oldParent.Bucket(oldkey)
	newBuck, err := newParent.CreateBucket(newkey)
	if err != nil {
		return err
	}

	err = oldBuck.ForEach(func(k, v []byte) error {
		if v == nil {
			// Nested bucket
			return moveBucket(oldBuck, newBuck, k, k)
		} else {
			// Regular value
			return newBuck.Put(k, v)
		}
	})
	if err != nil {
		return err
	}

	return oldParent.DeleteBucket(oldkey)
}
