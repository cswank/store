package site

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/shopify"
	"github.com/cswank/store/internal/templates"
	"github.com/nfnt/resize"
)

var (
	sizes = map[string]uint{
		"product.png": 360,
		"thumb.png":   200,
	}

	pages = []func([]link, categories) error{
		generateHome,
		generateContact,
		generateCart,
		generateSubcategories,
		generateProducts,
	}

	cfg config.Config
)

func Init(c config.Config) {
	cfg = c
}

type categories map[string]map[string][]string

func Generate() error {
	cats, err := getCategories()
	if err != nil {
		return err
	}

	links := getNavbarLinks(cats)

	for _, f := range pages {
		if err := f(links, cats); err != nil {
			return err
		}
	}

	return nil
}

func generateHome(links []link, cats categories) error {
	p := page{
		Links: links,
		//Shopify: shopify,
		Name: cfg.StoreName,
	}

	f, err := os.Create("index.html")
	if err != nil {
		return err
	}
	defer f.Close()
	return templates.Get("index.html").ExecuteTemplate(f, "base", p)
}

type contactPage struct {
	page
	Captcha        bool
	CaptchaSiteKey string
	ShowMessage    bool
}

func generateContact(links []link, cats categories) error {
	p := contactPage{
		page: page{
			Links:   links,
			Name:    cfg.StoreName,
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
		},
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
		Captcha:        true,
	}

	if !exists("contact") {
		if err := os.Mkdir("contact", 0700); err != nil {
			return err
		}
	}

	f, err := os.Create("contact/index.html")
	if err != nil {
		return err
	}
	defer f.Close()

	return templates.Get("contact.html").ExecuteTemplate(f, "base", p)
}

type cartPage struct {
	page
	Price string
}

func generateCart(links []link, cats categories) error {
	p := cartPage{
		page: page{
			Links: links,
			//Shopify: shopify,
			Name: cfg.StoreName,
		},
		Price: cfg.DefaultPrice,
	}

	if !exists("cart") {
		if err := os.Mkdir("cart", 0700); err != nil {
			return err
		}
	}

	f, err := os.Create("cart/index.html")
	if err != nil {
		return err
	}
	defer f.Close()

	return templates.Get("cart.html").ExecuteTemplate(f, "base", p)
}

func generateSubcategories(links []link, cats categories) error {
	for cat, subcats := range cats {
		for subcat, prods := range subcats {
			if err := generateSubcategory(links, cat, subcat, prods); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateSubcategory(links []link, cat, subcat string, names []string) error {

	p := subCategoryPage{
		page: page{
			Links: links,
			//Shopify: shopify,
			Name: cfg.StoreName,
		},
		Products: getProducts(cat, subcat, names),
	}

	f, err := os.Create(filepath.Join("products", cat, subcat, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return templates.Get("subcategory.html").ExecuteTemplate(f, "base", p)
}

func getProducts(cat, subcat string, prods []string) []Product {
	out := make([]Product, len(prods))
	for i, name := range prods {
		out[i] = Product{
			Title: name,
			Image: fmt.Sprintf("/products/%s/%s/%s/thumb.png", cat, subcat, name),
			Link:  fmt.Sprintf("/products/%s/%s/%s/index.html", cat, subcat, name),
			Price: cfg.DefaultPrice,
		}
	}
	return out
}

func generateProducts(links []link, cats categories) error {
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

	p := NewProduct(name, cat, subcat)

	var err error
	p.ID, err = shopify.Create(p.Title, p.Cat, cfg.DefaultPrice)
	if err != nil {
		return err
	}

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

	return templates.Get("product.html").ExecuteTemplate(f, "base", page)
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

type Product struct {
	Title       string
	Cat         string
	Subcat      string
	Price       string
	Total       string
	Quantity    int
	Description string
	ID          string
	Image       string
	Link        string
}

func NewProduct(title, cat, subcat string) Product {
	return Product{
		Title:    title,
		Cat:      cat,
		Subcat:   subcat,
		Price:    cfg.DefaultPrice,
		Quantity: 1,
	}
}

type subCategoryPage struct {
	page
	Products []Product
}

type productPage struct {
	page
	Product Product
}

func existingProduct(pth string) bool {
	item := filepath.Base(pth)
	return item == "thumb.png" || item == "product.png"
}

func getCategories() (categories, error) {
	c := categories{}
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

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
