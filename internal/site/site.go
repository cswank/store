package site

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

var (
	sizes = map[string]uint{
		"product.png": 360,
		"thumb.png":   200,
	}
)

type categories map[string]map[string][]string

func Generate() error {
	cats, err := getCategories()
	if err != nil {
		return err
	}

	links := getNavbarLinks(cats)

	for cat, m := range cats {
		for subcat, names := range m {
			for _, name := range names {
				if err := generateProduct(links, cat, subcat, name); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func generateProduct(links []link, cat, subcat, name string) error {

	p := newProduct(name, cat, subcat)

	page := productPage{
		page: page{
			Links: links,
			//Shopify:     shopify,
			Name:        name,
			Stylesheets: []string{"/css/product.css"},
		},
		Product: p,
	}

	dir := filepath.Join("products", cat, subcat, name)
	addImages(dir)

	f, err := os.Create(filepath.Join(dir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return templates["product.html"].template.ExecuteTemplate(f, "base", page)
}

func addImages(dir string) error {
	pth := filepath.Join(dir, "image.png")
	f, err := os.Open(pth)
	if err != nil {
		return err
	}

	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return err
	}

	for _, size := range []string{"product.png", "thumb.png"} {
		err := resizeImage(img, dir, size)
		if err != nil {
			return err
		}
	}
	return nil
}

func resizeImage(img image.Image, dir, name string) error {
	px := sizes[name]
	m := resize.Resize(px, 0, img, resize.Lanczos3)

	pth := filepath.Join(dir, name)
	f, err := os.Create(pth)
	if err != nil {
		return err
	}

	defer f.Close()
	return png.Encode(f, m)
}

type shopifyAPI struct {
	APIKey string
	Domain string
}

type page struct {
	Shopify     shopifyAPI
	Admin       bool
	Links       []link
	Scripts     []string
	Stylesheets []string
	Name        string
}

type product struct {
	Title       string
	Cat         string
	Subcat      string
	Price       string
	Quantity    int
	Description string
	ID          string
}

func newProduct(title, cat, subcat string) product {
	return product{
		Title:    title,
		Cat:      cat,
		Subcat:   subcat,
		Price:    cfg.DefaultPrice,
		Quantity: 1,
	}
}

type productPage struct {
	page
	Product product
}

func getCategories() (categories, error) {
	c := categories{}
	err := filepath.Walk("./products", func(pth string, info os.FileInfo, err error) error {
		if strings.HasSuffix(pth, ".png") {
			parts := strings.Split(pth, "/")
			if len(parts) != 5 {
				return fmt.Errorf("invalid path to image: %s", pth)
			}

			cat := parts[1]
			subcat := parts[2]
			name := parts[3]
			m, ok := c[cat]
			if !ok {
				m = map[string][]string{}
			}

			s, ok := m[subcat]
			if !ok {
				s = []string{}
			}

			s = append(s, name)
			m[subcat] = s
			c[cat] = m
		}
		return nil
	})
	return c, err
}
