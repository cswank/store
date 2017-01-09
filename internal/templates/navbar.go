package templates

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	lock     sync.Mutex
	links    []Link
	products map[string]map[string][]Product
	cats     Categories
)

func InitProducts() {
	products = getAllProducts()
}

type Page struct {
	Links       []Link
	Scripts     []string
	Stylesheets []string
	Name        string
}

type Link struct {
	Name     string
	Link     string
	Style    string
	Children []Link
}

type Categories map[string]map[string][]string

func GetLinks() []Link {
	lock.Lock()
	defer lock.Unlock()
	return links
}

func GetCategories() Categories {
	lock.Lock()
	defer lock.Unlock()
	return cats
}

func initLinks() {
	var err error
	cats, err = getCategories()
	if err != nil {
		log.Fatal(err)
	}

	links = getNavbarLinks(cats)
}

func getNavbarLinks(cats Categories) []Link {
	return []Link{
		{Name: "Home", Link: "/"},
		{Name: "Shop", Link: "/", Children: getShoppingLinks(cats)},
		{Name: "Contact", Link: "/contact"},
		{Name: "Wholesale", Link: "/wholesale"},
		{Name: "Cart", Link: "/cart"},
	}
}

func getShoppingLinks(cats Categories) []Link {
	var l []Link

	for cat, subcats := range cats {
		l = append(l, getSubcatLinks(cat, subcats)...)
	}

	return l
}

func getSubcatLinks(cat string, subcats map[string][]string) []Link {

	l := make([]Link, len(subcats))

	var i int
	for subcat := range subcats {
		l[i] = Link{
			Link: fmt.Sprintf("/products/%s/%s", cat, subcat),
			Name: subcat,
		}
		i++
	}

	return l
}

func getCategories() (Categories, error) {
	c := Categories{}
	err := filepath.Walk("./products", func(pth string, info os.FileInfo, err error) error {
		if strings.HasSuffix(pth, ".png") {
			if existingProduct(pth) {
				return nil
			}

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

func existingProduct(pth string) bool {
	item := filepath.Base(pth)
	return item == "thumb.png" || item == "product.png"
}

func GetProducts(cat, subcat string, prods []string) []Product {
	lock.Lock()
	defer lock.Unlock()
	return products[cat][subcat]
}

func getAllProducts() map[string]map[string][]Product {
	m := map[string]map[string][]Product{}
	for cat, subcats := range cats {
		a := map[string][]Product{}
		for subcat, prods := range subcats {
			a[subcat] = GetProductsForSubcat(cat, subcat, prods)
		}
		m[cat] = a
	}
	return m
}

func GetProductsForSubcat(cat, subcat string, prods []string) []Product {
	out := make([]Product, len(prods))
	for i, name := range prods {
		pth := filepath.Join("products", cat, subcat, name, "id.txt")
		d, err := ioutil.ReadFile(pth)
		if err != nil {
			log.Fatal("could not read product id: ", pth)
		}

		out[i] = Product{
			ID:    string(d),
			Title: name,
			Image: fmt.Sprintf("/products/%s/%s/%s/product.png", cat, subcat, name),
			Thumb: fmt.Sprintf("/products/%s/%s/%s/thumb.png", cat, subcat, name),
			Link:  fmt.Sprintf("/products/%s/%s/%s/index.html", cat, subcat, name),
			Price: cfg.DefaultPrice,
		}
	}
	return out
}
