package store

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/boltdb/bolt"
)

var (
	search ItemSearcher
)

type Searcher func(Item) bool
type Booler func(Item, ...Searcher) bool

type ItemSearcher interface {
	Search(Booler, ...Searcher) []Item
	Update(*Item)
}

type memorySearcher struct {
	items map[string]Item
	lock  sync.Mutex
}

func initSearch() {
	items := map[string]Item{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var i Item
			if err := json.Unmarshal(v, &i); err != nil {
				return err
			}
			items[i.ID] = i
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	search = &memorySearcher{
		items: items,
	}
}

func or(item Item, searchers ...Searcher) bool {
	for _, s := range searchers {
		if s(item) {
			return true
		}
	}
	return false
}

func and(item Item, searchers ...Searcher) bool {
	for _, s := range searchers {
		if !s(item) {
			return false
		}
	}
	return true
}

func (m *memorySearcher) Update(item *Item) {
	m.lock.Lock()
	m.items[item.ID] = *item
	m.lock.Unlock()
}

func (m *memorySearcher) Search(andOr Booler, searchers ...Searcher) []Item {
	var items []Item
	m.lock.Lock()
	for _, item := range items {
		if andOr(item, searchers...) {
			items = append(items, item)
		}
	}
	m.lock.Unlock()
	return items
}

func searchCategory(cat string) func(Item) bool {
	return func(item Item) bool {
		return cat == item.Category
	}
}
