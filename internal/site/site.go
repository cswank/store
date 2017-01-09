package site

import (
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"

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

	pages = []func([]templates.Link, templates.Categories) error{
		generateProducts,
		generateHome,
		generateContact,
		generateCart,
		generateThumbs,
	}

	cfg     config.Config
	shopAPI shopifyAPI
)

func Init(c config.Config) {
	cfg = c
	shopAPI = shopifyAPI{
		APIKey: cfg.ShopifyJSKey,
		Domain: cfg.ShopifyDomain,
	}
}

func Generate() error {
	links := templates.GetLinks()
	cats := templates.GetCategories()

	for _, f := range pages {
		if err := f(links, cats); err != nil {
			return err
		}
	}

	return nil
}

func generateHome(links []templates.Link, cats templates.Categories) error {
	p := templates.Page{
		Links: templates.GetLinks(),
		Name:  cfg.StoreName,
	}

	f, err := os.Create("index.html")
	if err != nil {
		return err
	}
	defer f.Close()
	return templates.Get("index.html").ExecuteTemplate(f, "base", p)
}

type formPage struct {
	templates.Page
	Captcha        bool
	CaptchaSiteKey string
	ShowMessage    bool
}

func generateContact(links []templates.Link, cats templates.Categories) error {
	p := formPage{
		Page: templates.Page{
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
	templates.Page
	Price   string
	Shopify shopifyAPI
}

func generateCart(links []templates.Link, cats templates.Categories) error {
	p := cartPage{
		Page: templates.Page{
			Links: links,
			Name:  cfg.StoreName,
		},
		Shopify: shopAPI,
		Price:   cfg.DefaultPrice,
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

func generateThumbs(links []templates.Link, cats templates.Categories) error {
	for cat, subcats := range cats {
		for subcat, prods := range subcats {
			if err := generateThumbPage(links, cat, subcat, prods); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateThumbPage(links []templates.Link, cat, subcat string, names []string) error {
	p := templates.ThumbPage{
		Page: templates.Page{
			Links: links,
			Name:  cfg.StoreName,
		},
		Products: templates.GetProductsForSubcat(cat, subcat, names),
	}

	f, err := os.Create(filepath.Join("products", cat, subcat, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return templates.Get("thumbs.html").ExecuteTemplate(f, "base", p)
}

func generateProducts(links []templates.Link, cats templates.Categories) error {
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

func generateProduct(links []templates.Link, cat, subcat, name string) error {

	p := templates.NewProduct(name, cat, subcat)

	var err error
	p.ID, err = shopify.Create(p.Title, p.Cat, cfg.DefaultPrice)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join("products", cat, subcat, name, "id.txt"))
	if err != nil {
		return err
	}

	if _, err := f.Write([]byte(p.ID)); err != nil {
		return err
	}
	f.Close()

	page := productPage{
		Page: templates.Page{
			Links:       links,
			Name:        name,
			Stylesheets: []string{"/css/product.css"},
		},
		Product: p,
	}

	dir := filepath.Join("products", cat, subcat, name)
	if err := addImages(dir); err != nil {
		return err
	}

	f, err = os.Create(filepath.Join(dir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := addImageToShopify(cat, subcat, name, p.ID); err != nil {
		return err
	}

	return templates.Get("product.html").ExecuteTemplate(f, "base", page)
}

func addImageToShopify(cat, subcat, name, id string) error {
	pth := filepath.Join("products", cat, subcat, name, "product.png")
	img, err := ioutil.ReadFile(pth)
	if err != nil {
		return err
	}

	return shopify.AddImage(id, img)
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

type productPage struct {
	templates.Page
	Product templates.Product
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
