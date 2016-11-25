package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/ventu-io/go-shortid"
)

type Item struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	SubCategory string   `json:"sub_category"`
	Page        int      `json:"page"`
	Keywords    []string `json:"keywords"`

	//Price in cents
	Price float32 `json:"price"`
}

func NewItem() (*Item, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	return &Item{
		ID: id,
	}, nil
}

func GetSubCatetory(cat, subCat string, page int) ([]Item, error) {
	var items []Item
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items")).Bucket([]byte(cat)).Bucket([]byte(subCat)).Bucket([]byte(fmt.Sprintf("%d", page)))
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
	return items, err
}

func (i *Item) Fetch() error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		v := b.Get([]byte(i.ID))
		if len(v) == 0 {
			return ErrNotFound
		}
		return json.Unmarshal(v, i)
	})
}

func (i *Item) Save() error {
	if i.ID == "" {
		return errors.New("can't save item with no ID")
	}

	return db.Update(func(tx *bolt.Tx) error {
		d, _ := json.Marshal(i)
		b := tx.Bucket([]byte("items"))
		var err error
		b, err = i.createBuckets(b)
		if err != nil {
			return err
		}
		return b.Put([]byte(i.ID), d)
	})
}

func (i *Item) createBuckets(b *bolt.Bucket) (*bolt.Bucket, error) {
	var err error
	for _, name := range [][]byte{[]byte(i.Category), []byte(i.SubCategory), []byte(strconv.FormatInt(int64(i.Page), 10))} {
		if b, err = b.CreateBucketIfNotExists(name); err != nil {
			return nil, err
		}
	}
	return b, nil
}
