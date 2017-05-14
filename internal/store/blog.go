package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"strings"
	"time"
)

type Blog struct {
	Title string    `json:"title" schema:"title"`
	Date  time.Time `json:"date" schema:"date"`
	Body  string    `json:"body" schema:"body"`

	image io.Reader
}

func CurrentBlog() (Blog, error) {
	var key string
	err := db.Get([]Query{NewQuery(Key("current"), Buckets("blogs"))}, func(_, val []byte) error {
		key = string(val)
		return nil
	})

	if err != nil {
		return Blog{}, err
	}

	if key == "" {
		return Blog{}, nil
	}

	return GetBlog(key)

}

func GetBlog(key string) (Blog, error) {
	var b Blog
	return b, db.Get([]Query{NewQuery(Key(key), Buckets("blogs"))}, func(_, val []byte) error {
		return json.Unmarshal(val, &b)
	})
}

type BlogKey struct {
	Date  string
	Title string
	ID    string
}

func Blogs() ([]BlogKey, error) {
	var blogs []BlogKey
	err := db.GetAll(Query{Buckets: [][]byte{[]byte("blogs")}}, func(key, _ []byte) error {
		k := string(key)
		i := strings.Index(k, ":")
		if i > -1 {
			bk := BlogKey{ID: k, Date: k[:i], Title: k[i+1:]}
			blogs = append(blogs, bk)
		}
		return nil
	})

	if err == ErrNotFound {
		return blogs, nil
	}
	return blogs, err
}

func (b *Blog) Key() string {
	return fmt.Sprintf("%s:%s", b.Date.Format("2006-01-02"), b.Title)
}

func (b *Blog) Fetch() error {
	return db.Get([]Query{{Key: []byte(b.Key()), Buckets: [][]byte{[]byte("blogs")}}}, func(key, val []byte) error {
		return json.Unmarshal(val, &b)
	})
}

func (b *Blog) Update(b2 Blog, img io.Reader) error {
	b.Body = b2.Body
	if b.Title != b2.Title || b.Date != b2.Date {

		q := []Query{
			NewQuery(Key(b.Key()), Buckets("blogs")),
		}

		if img == nil {
			// get old image data and delete it from db
			db.Get([]Query{NewQuery(Key(b.Key()), Buckets("images", "blogs"))}, func(_, val []byte) error {
				img = bytes.NewBuffer(val)
				return nil
			})
			q = append(q, NewQuery(Key(b.Key()), Buckets("images", "blogs")))
		}

		if err := db.Delete(q); err != nil {
			return err
		}

		b.Title = b2.Title
		b.Date = b2.Date
	}

	return b.doSave(img)
}

func (b *Blog) Save(img io.Reader) error {
	b.Date = time.Now()
	return b.doSave(img)
}

func (b *Blog) doSave(img io.Reader) error {
	d, err := json.Marshal(b)
	if err != nil {
		return err
	}

	q := []Query{
		NewQuery(Key(b.Key()), Val(d), Buckets("blogs")),
		NewQuery(Key("current"), Val([]byte(b.Key())), Buckets("blogs")),
	}

	if img != nil {
		var err error
		q, err = addBlogImage(img, b.Key(), q)
		if err != nil {
			return err
		}
	}

	return db.Put(q)
}

func GetBlogImage(key string) ([]byte, error) {
	q := []Query{NewQuery(Key(key), Buckets("images", "blogs"))}
	var img []byte
	return img, db.Get(q, func(k, v []byte) error {
		img = v
		return nil
	})
}

func addBlogImage(r io.Reader, blog string, q []Query) ([]Query, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	d, err := resizeImage(img, uint(full))
	if err != nil {
		return nil, err
	}

	return append(
		q,
		NewQuery(Key(blog), Val(d), Buckets("images", "blogs")),
	), nil
}

func (b *Blog) Delete() error {
	return db.Delete([]Query{
		NewQuery(Buckets("blogs"), Key(b.Key())),
		NewQuery(Buckets("images", "blogs"), Key(b.Key())),
	})
}
