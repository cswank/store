package mock

import (
	"bytes"
	"net/http"

	"github.com/cswank/store/internal/store"
)

type Result struct {
	Key []byte
	Val []byte
}

type DB struct {
	i       int
	errors  []error
	buckets map[string][]Result
	Rows    []store.Query
}

func NewDB(buckets map[string][]Result, errors []error) *DB {
	return &DB{
		buckets: buckets,
		errors:  errors,
		Rows:    []store.Query{},
	}
}

func (d *DB) Put(rows []store.Query) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, rows...)
	return err
}

func (d *DB) AddBucket(row store.Query) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, row)
	return err
}

func (d *DB) GetBackup(w http.ResponseWriter) error {
	return nil
}

func (d *DB) RenameBucket(dst, src store.Query) error {
	return nil
}

func (d *DB) Get(rows []store.Query, f func([]byte, []byte) error) error {
	for _, r := range rows {
		d.Rows = append(d.Rows, r)
		err := d.errors[d.i]
		if err != nil {
			return err
		}
		k := string(bytes.Join(r.Buckets, []byte(" ")))
		results := d.buckets[k]
		for _, res := range results {
			if bytes.Equal(r.Key, res.Key) {
				if err := f(res.Key, res.Val); err != nil {
					return err
				}
			}
		}
		d.i++
	}
	return nil
}

func (d *DB) GetAll(row store.Query, f func(key, val []byte) error) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, row)
	k := string(bytes.Join(row.Buckets, []byte(" ")))
	results := d.buckets[k]
	for _, res := range results {
		if err := f(res.Key, res.Val); err != nil {
			return err
		}
	}
	return err
}

func (d *DB) Delete(rows []store.Query) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, rows...)
	return err
}

func (d *DB) DeleteAll(bucket []byte) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, store.Query{Buckets: [][]byte{bucket}})
	return err
}
