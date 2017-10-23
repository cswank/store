package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"

	"github.com/cswank/store/internal/shopify"
	"github.com/nfnt/resize"
)

const (
	full  = 660
	thumb = 260
)

var (
	sizeNames = map[int]string{
		full:  "image.png",
		thumb: "thumb.png",
	}

	//ErrExists indicates you are trying to add an item in the db that already exists
	ErrExists = errors.New("product already exists")
)

type Price struct {
	Price          string `json:"price"`
	WholesalePrice string `json:"wholesale_price"`
}

func GetPrice(name string) (Price, error) {
	q := []Query{NewQuery(Buckets("products", name), Key("_price_"))}
	var p Price
	return p, db.Get(q, func(key, val []byte) error {
		return json.Unmarshal(val, &p)
	})
}

func AddCategory(name, p, w string) error {
	row := NewQuery(Buckets("products"), Key(name))
	if err := db.AddBucket(row); err != nil {
		return err
	}

	d, err := json.Marshal(Price{Price: p, WholesalePrice: w})
	if err != nil {
		return err
	}

	return db.Put([]Query{
		NewQuery(Buckets("products", name), Key("_price_"), Val(d)),
	})
}

func RenameCategory(old, name string) error {
	src := NewQuery(Buckets("products"), Key(old))
	dst := NewQuery(Buckets("products"), Key(name))
	return db.RenameBucket(src, dst)
}

func DeleteCategory(cat string) error {
	rows := []Query{NewQuery(Buckets("products", cat))}
	return db.Delete(rows)
}

func AddSubcategory(cat, name string) error {
	row := NewQuery(Buckets("products", cat), Key(name))
	return db.AddBucket(row)
}

func RenameSubcategory(cat, old, name string) error {
	src := NewQuery(Buckets("products", cat), Key(old))
	dst := NewQuery(Buckets("products", cat), Key(name))
	return db.RenameBucket(src, dst)
}

func DeleteSubcategory(cat, subcat string) error {
	rows := []Query{NewQuery(Buckets("products", cat, subcat))}
	return db.Delete(rows)
}

func GetCategories() ([]string, error) {
	var cats []string
	q := NewQuery(Buckets("products"))
	return cats, db.GetAll(q, func(key, val []byte) error {
		cats = append(cats, string(key))
		return nil
	})
}

func GetSubCategories(cat string) ([]string, error) {
	var cats []string
	q := NewQuery(Buckets("products", cat))
	return cats, db.GetAll(q, func(key, val []byte) error {
		cat := string(key)
		if cat != "_price_" {
			cats = append(cats, cat)
		}
		return nil
	})
}

func GetProductTitles(cat, subcat string) ([]string, error) {
	var products []string
	q := NewQuery(Buckets("products", cat, subcat))
	return products, db.GetAll(q, func(key, val []byte) error {
		products = append(products, string(key))
		return nil
	})
}

func GetProducts(cat, subcat string, opts ...func(*Product)) ([]Product, error) {
	var products []Product
	q := NewQuery(Buckets("products", cat, subcat))
	return products, db.GetAll(q, func(key, val []byte) error {
		var p Product
		err := json.Unmarshal(val, &p)
		if err != nil {
			return err
		}
		p.Title = string(key)

		for _, o := range opts {
			o(&p)
		}

		products = append(products, p)
		return nil
	})
}

func GetBackup(w http.ResponseWriter) error {
	return db.GetBackup(w)
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

	image io.Reader
}

func NewProduct(title, cat, subcat string, opts ...func(*Product)) *Product {
	p := &Product{
		Title:  title,
		Cat:    cat,
		Subcat: subcat,
		Price:  cfg.DefaultPrice,
	}

	for _, o := range opts {
		o(p)
	}
	return p
}

func ProductPrice(price string) func(*Product) {
	return func(p *Product) {
		p.Price = price
	}
}

func ProductDescription(d string) func(*Product) {
	return func(p *Product) {
		p.Description = d
	}
}

func ProductImage(r io.Reader) func(*Product) {
	return func(p *Product) {
		p.image = r
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
	p.Description = p2.Description

	if p2.Subcat != p.Subcat {
		if err := p.move(p2.Subcat); err != nil {
			return err
		}
	}

	if p2.Title != p.Title {
		//rename images
		//delete old key
		//save new key
	}

	buckets := []string{
		"products",
		p.Cat,
		p.Subcat,
	}

	d, err := json.Marshal(p)
	if err != nil {
		return err
	}
	rows := []Query{NewQuery(Key(p.Title), Val(d), Buckets(buckets...))}

	if p2.image != nil {
		var imgQueries []Query
		_, imgQueries, err = addImage(p2.image, p.Title, "products")
		if err != nil {
			return err
		}
		rows = append(rows, imgQueries...)
	}

	return db.Put(rows)
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

func (p *Product) query() []Query {
	d, _ := json.Marshal(p)
	return []Query{
		NewQuery(Key(p.Title), Val(d), Buckets("products", p.Cat, p.Subcat)),
	}
}

func (p *Product) imageQuery() []Query {
	return []Query{
		NewQuery(Buckets("images", "products", p.Title)),
	}
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

func (p *Product) getSubcat(rows []Query) ([]Query, error) {
	q := NewQuery(Buckets("products", p.Cat, p.Subcat))
	err := db.GetAll(q, func(key, val []byte) error {
		id := string(key)
		if id == p.Title {
			return ErrExists
		}
		return nil
	})

	if err != nil {
		return []Query{}, err
	}

	buckets := []string{
		"products",
		p.Cat,
		p.Subcat,
	}

	d, err := json.Marshal(p)
	if err != nil {
		return []Query{}, err
	}
	r := NewQuery(Key(p.Title), Val(d), Buckets(buckets...))
	rows = append(rows, r)
	return rows, nil
}

func addImage(r io.Reader, name, bucket string) ([]byte, []Query, error) {
	var imgData []byte
	img, err := png.Decode(r)
	if err != nil {
		return nil, nil, err
	}

	rows := make([]Query, 2)
	for i, s := range []int{full, thumb} {
		d, err := resizeImage(img, uint(s))
		if err != nil {
			return nil, nil, err
		}

		if i == 0 {
			imgData = d
		}

		r := NewQuery(Key(sizeNames[s]), Val(d), Buckets("images", bucket, name))
		rows[i] = r
	}
	return imgData, rows, nil
}

func resizeImage(img image.Image, size uint) ([]byte, error) {
	m := resize.Resize(size, 0, img, resize.Bilinear)
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
