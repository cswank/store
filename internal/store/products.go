package store

import (
	"bytes"
	"image/jpeg"
	"io"

	"github.com/cswank/store/internal/shopify"
	"github.com/nfnt/resize"
)

const (
	full  = 360
	thumb = 200
)

var sizeNames = map[int]string{
	full:  "",
	thumb: "-thumb",
}

func GetCategories() ([]string, error) {
	return nil, nil
}

func GetSubCategories(cat string) ([]string, error) {
	return nil, nil
}

func GetProducts(cat, subcat string) ([]string, error) {
	return nil, nil
}

func AddProduct(cat, subcat, title string, r io.Reader) error {
	id, err := shopify.CreateProduct(title, cat)
	if err != nil {
		return err
	}

	if err := addImages(title, r); err != nil {
		return err
	}

	//add shopify id
	//add to products -> shopifyid
	//add to categories -> subcat: title
	return nil
}

func addImages(id string, r io.Reader) error {
	img, err := jpeg.Decode(r)
	if err != nil {
		return err
	}

	for _, s := range []int{full, thumb} {
		m := resize.Resize(uint(s), 0, img, resize.Lanczos3)

		var buf bytes.Buffer

		jpeg.Encode(&buf, m, nil)

		if err := db.Put([]byte(id+sizeNames[s]), buf.Bytes(), []byte("images")); err != nil {
			return err
		}
	}
	return nil
}

/*
products
   Cards
      Happy Birthday
        uid1 -> shopify product id
        uid2 -> shopify product id
      Anniversary
        uid2 -> shopify product id
        uid4 -> shopify product id

images
  "uid1 thumb" -> thumb
  "uid1 image" -> image
*/
