package mock

import (
	"bytes"
	"net/http"

	"github.com/cswank/store/internal/storage"
)

type Result struct {
	Key []byte
	Val []byte
}

type DB struct {
	i       int
	errors  []error
	buckets map[string][]Result
	Rows    []storage.Row
}

func NewDB(buckets map[string][]Result, errors []error) *DB {
	return &DB{
		buckets: buckets,
		errors:  errors,
		Rows:    []storage.Row{},
	}
}

func (d *DB) Put(rows []storage.Row) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, rows...)
	return err
}

func (d *DB) AddBucket(row storage.Row) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, row)
	return err
}

func (d *DB) GetBackup(w http.ResponseWriter) error {
	return nil
}

func (d *DB) RenameBucket(dst, src storage.Row) error {
	return nil
}

func (d *DB) Get(rows []storage.Row, f func([]byte, []byte) error) error {
	for _, r := range rows {
		d.Rows = append(d.Rows, r)
		err := d.errors[d.i]
		d.i++
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
	}
	return nil
}

func (d *DB) GetAll(row storage.Row, f func(key, val []byte) error) error {
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

func (d *DB) Delete(rows []storage.Row) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, rows...)
	return err
}

func (d *DB) DeleteAll(bucket []byte) error {
	err := d.errors[d.i]
	d.i++
	d.Rows = append(d.Rows, storage.Row{Buckets: [][]byte{bucket}})
	return err
}
