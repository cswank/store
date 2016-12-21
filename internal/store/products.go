package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"

	"github.com/cswank/store/internal/shopify"
	"github.com/nfnt/resize"
)

const (
	full  = 360
	thumb = 200
)

var (
	sizeNames = map[int]string{
		full:  "image.png",
		thumb: "thumb.png",
	}

	//ErrExists indicates you are trying to add an item in the db that already exists
	ErrExists = errors.New("product already exists")
)

func AddCategory(name string) error {
	row := Row{Buckets: [][]byte{[]byte("products")}, Key: []byte(name)}
	return db.AddBucket(row)
}

func RenameCategory(old, name string) error {
	src := Row{Buckets: [][]byte{[]byte("products")}, Key: []byte(old)}
	dst := Row{Buckets: [][]byte{[]byte("products")}, Key: []byte(name)}
	return db.RenameBucket(src, dst)
}

func AddSubcategory(cat, name string) error {
	row := Row{Buckets: [][]byte{[]byte("products"), []byte(cat)}, Key: []byte(name)}
	return db.AddBucket(row)
}

func RenameSubcategory(cat, old, name string) error {
	src := Row{Buckets: [][]byte{[]byte("products"), []byte(cat)}, Key: []byte(old)}
	dst := Row{Buckets: [][]byte{[]byte("products"), []byte(cat)}, Key: []byte(name)}
	return db.RenameBucket(src, dst)
}

func GetCategories() ([]string, error) {
	var cats []string
	q := Row{Buckets: [][]byte{[]byte("products")}}
	return cats, db.GetAll(q, func(key, val []byte) error {
		cats = append(cats, string(key))
		return nil
	})
}

func GetSubCategories(cat string) ([]string, error) {
	var cats []string
	q := Row{Buckets: [][]byte{[]byte("products"), []byte(cat)}}
	return cats, db.GetAll(q, func(key, val []byte) error {
		cats = append(cats, string(key))
		return nil
	})
}

func GetProducts(cat, subcat string) ([]string, error) {
	var products []string
	q := Row{Buckets: [][]byte{[]byte("products"), []byte(cat), []byte(subcat)}}
	return products, db.GetAll(q, func(key, val []byte) error {
		products = append(products, string(key))
		return nil
	})
}

type Product struct {
	Title       string `json:"-"`
	Cat         string `json:"-"`
	Subcat      string `json:"-"`
	Price       string `json:"-"`
	Total       string `json:"-"`
	Quantity    int    `json:"-"`
	Description string `json:"description"`
	ID          string `json:"id"`
}

func NewProduct(title, cat, subcat, description string) *Product {
	return &Product{
		Title:       title,
		Cat:         cat,
		Subcat:      subcat,
		Description: description,
		Price:       DefaultPrice,
	}
}

func (p *Product) Fetch() error {
	if p.Title == "" {
		return errors.New("product title must be set")
	}

	return db.Get(p.query(), func(key, val []byte) error {
		return json.Unmarshal(val, p)
	})
}

func (p *Product) Update(p2 *Product) error {
	if p2.Subcat != p.Subcat {
		if err := p.move(p2.Subcat); err != nil {
			return err
		}
	}

	if p2.Title == p.Title {
		return nil
	}

	//TODO enable rename
	return nil
}

func (p *Product) move(dst string) error {
	if err := p.Fetch(); err != nil {
		return err
	}

	if err := db.Delete(p.query()); err != nil {
		return err
	}

	q := p.query()
	b := q[0].Buckets
	b[2] = []byte(dst)
	q[0].Buckets = b

	return db.Put(q)
}

func (p *Product) Delete() error {
	if err := shopify.Delete(p.ID); err != nil {
		return err
	}
	q := append(p.query(), p.imageQuery()...)
	return db.Delete(q)
}

func (p *Product) query() []Row {
	d, _ := json.Marshal(p)
	return []Row{{Key: []byte(p.Title), Val: d, Buckets: [][]byte{[]byte("products"), []byte(p.Cat), []byte(p.Subcat)}}}
}

func (p *Product) imageQuery() []Row {
	return []Row{{Buckets: [][]byte{[]byte("images"), []byte("products"), []byte(p.Title)}}}
}

func (p *Product) Add(r io.Reader) error {
	id, err := shopify.Create(p.Title, p.Cat, cfg.DefaultPrice)
	if err != nil {
		return err
	}
	p.ID = id

	img, rows, err := addImage(r, p.Title, "products")
	if err != nil {
		return err
	}

	rows, err = p.getSubcat(rows)
	if err != nil {
		return err
	}

	if err := db.Put(rows); err != nil {
		return err
	}

	return shopify.AddImage(id, img)
}

func (p *Product) getSubcat(rows []Row) ([]Row, error) {
	var ids []string
	q := Row{Buckets: [][]byte{[]byte("products"), []byte(p.Cat), []byte(p.Subcat)}}
	err := db.GetAll(q, func(key, val []byte) error {
		id := string(key)
		if id == p.Title {
			return ErrExists
		}
		ids = append(ids, p.ID)
		return nil
	})

	if err != nil {
		return []Row{}, err
	}

	buckets := [][]byte{
		[]byte("products"),
		[]byte(p.Cat),
		[]byte(p.Subcat),
	}

	d, err := json.Marshal(p)
	if err != nil {
		return []Row{}, err
	}
	r := Row{Key: []byte(p.Title), Val: d, Buckets: buckets}
	rows = append(rows, r)
	return rows, nil
}

func addImage(r io.Reader, name, bucket string) ([]byte, []Row, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, nil, err
	}

	var imgData []byte
	rows := make([]Row, 2)
	for i, s := range []int{full, thumb} {
		d, err := resizeImage(img, uint(s))
		if err != nil {
			return nil, nil, err
		}

		if i == 0 {
			imgData = d
		}

		r := Row{Key: []byte(sizeNames[s]), Val: d, Buckets: [][]byte{[]byte("images"), []byte(bucket), []byte(name)}}
		rows[i] = r
	}
	return imgData, rows, nil
}

func resizeImage(img image.Image, size uint) ([]byte, error) {
	m := resize.Resize(size, 0, img, resize.Lanczos3)
	var buf bytes.Buffer
	if err := png.Encode(&buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/*
products
   Cards
      Happy Birthday
        uid1: shopify product id
        uid2: shopify product id
      Anniversary
        uid2: shopify product id
        uid4: shopify product id

images
  "uid1"
      thumb
      image
*/
