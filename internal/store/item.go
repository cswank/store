package store

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/ventu-io/go-shortid"
)

type Item struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	SubCategory string   `json:"sub_category"`
	Keywords    []string `json:"keywords"`

	//Price in cents
	Price int `json:"price"`
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

func GetCategory(category string) []Item {
	return search.Search(or, searchCategory(category))
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

	err := db.Update(func(tx *bolt.Tx) error {
		d, _ := json.Marshal(i)
		b := tx.Bucket([]byte("items"))
		return b.Put([]byte(i.ID), d)
	})

	if err == nil {
		search.Update(i)
	}

	return err
}
