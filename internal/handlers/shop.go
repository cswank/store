package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/mux"
)

type shopifyAPI struct {
	APIKey string
	Domain string
}

var (
	shopifyKey shopifyAPI
)

type shopPage struct {
	page
	Shopify    shopifyAPI
	Categories []string
}

func Shop(w http.ResponseWriter, req *http.Request) error {
	cats, err := store.GetCategories()
	if err != nil {
		return err
	}

	p := shopPage{
		Shopify:    shopifyKey,
		Categories: cats,
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Shopify: shopifyKey,
		},
	}
	return templates.Get("shop.html").ExecuteTemplate(w, "base", p)
}

type cartPage struct {
	page
	Price             string
	DiscountCode      string
	UnderConstruction bool
}

func Cart(w http.ResponseWriter, req *http.Request) error {

	dc := ""
	if Wholesaler(req) {
		dc = cfg.DiscountCode
	}

	p := cartPage{
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Shopify: shopifyKey,
			Name:    name,
		},
		Price:             getPrice(req),
		DiscountCode:      dc,
		UnderConstruction: cfg.UnderConstruction,
	}
	return templates.Get("cart.html").ExecuteTemplate(w, "base", p)
}

func getPrice(req *http.Request) string {
	price := cfg.DefaultPrice
	if Wholesaler(req) {
		price = cfg.WholesalePrice
	}
	return price
}

func LineItem(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	price := getPrice(req)
	p := store.NewProduct(vars["title"], cat, subcat, store.ProductPrice(price))
	if err := p.Fetch(); err != nil {
		return err
	}

	vals := req.URL.Query()

	qs := vals.Get("quantity")
	if qs == "" {
		return fmt.Errorf("you must supply a quantity")
	}

	q, err := strconv.ParseInt(qs, 10, 64)
	if err != nil {
		return err
	}

	p.Quantity = int(q)
	pr, err := strconv.ParseFloat(price, 10)
	if err != nil {
		return err
	}

	t := float64(q) * pr
	p.Total = fmt.Sprintf("%.02f", t)
	return templates.Get("lineitem.html").ExecuteTemplate(w, "lineitem.html", p)
}

type categoryPage struct {
	page
	SubCategories []link
	Products      map[string][]product
}

func Category(w http.ResponseWriter, req *http.Request) error {
	cat, _, _ := getVars(req)

	subs, err := store.GetSubCategories(cat)
	if err != nil {
		return err
	}

	if len(subs) == 1 && subs[0] == "NOSUBCATEGORIES" {
		return showSubcategory(cat, subs[0], w, req)
	}

	links := getLinks(fmt.Sprintf("/shop/%s", cat), subs)
	products, err := getProductsFromLinks(cat, links, getPrice(req))
	if err != nil {
		return err
	}

	p := categoryPage{
		SubCategories: links,
		Products:      products,
		page: page{
			Admin:   Admin(req),
			Links:   getNavbarLinks(req),
			Shopify: shopifyKey,
			Name:    name,
		},
	}
	return templates.Get("category.html").ExecuteTemplate(w, "base", p)
}

func getProductsFromLinks(cat string, links []link, price string) (map[string][]product, error) {
	m := map[string][]product{}
	for _, l := range links {
		prods, err := store.GetProducts(cat, l.Name)
		if err != nil {
			return nil, err
		}
		m[l.Name] = getProducts(cat, l.Name, prods, price)
	}
	return m, nil
}

func getLinks(href string, names []string) []link {
	links := make([]link, len(names))
	for i, n := range names {
		links[i] = link{Name: n, Link: path.Join(href, n)}
	}
	return links
}

type subCategoryPage struct {
	page
	Products []product
}

func getProducts(cat, subcat string, prods []store.Product, price string) []product {
	out := make([]product, len(prods))
	for i, p := range prods {
		out[i] = product{
			Title: p.Title,
			Image: fmt.Sprintf("/shop/images/products/%s/thumb.png", p.Title),
			Link:  fmt.Sprintf("/shop/%s/%s/%s", cat, subcat, p.Title),
			Price: price,
			ID:    p.ID,
		}
	}
	return out
}

func getVars(req *http.Request) (string, string, map[string]string) {
	vars := mux.Vars(req)
	cat := vars["category"]
	subcat := vars["subcategory"]
	return cat, subcat, vars
}

func SubCategory(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, _ := getVars(req)
	return showSubcategory(cat, subcat, w, req)

}

func showSubcategory(cat, subcat string, w http.ResponseWriter, req *http.Request) error {
	prods, err := store.GetProducts(cat, subcat)
	if err != nil {
		return err
	}

	p := subCategoryPage{
		page: page{
			Admin:   Admin(req),
			Links:   getNavbarLinks(req),
			Shopify: shopifyKey,
			Name:    name,
		},
		Products: getProducts(cat, subcat, prods, getPrice(req)),
	}
	return templates.Get("subcategory.html").ExecuteTemplate(w, "base", p)
}

type productPage struct {
	page
	Product     store.Product
	Back        string
	BackText    string
	Subcategory string
}

type product struct {
	ID        string
	Title     string
	Image     string
	Link      string
	ProductID string
	Price     string
}

func GetProduct(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat, store.ProductPrice(getPrice(req)))
	if err := p.Fetch(); err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(p)
}

func Product(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat, store.ProductPrice(getPrice(req)))

	if err := p.Fetch(); err != nil {
		return err
	}

	p.Quantity = 1

	var back, backText string
	if subcat == "NOSUBCATEGORIES" {
		back = fmt.Sprintf("/shop/%s", cat)
		backText = cat
	} else {
		back = fmt.Sprintf("/shop/%s/%s", cat, subcat)
		backText = subcat
	}

	page := productPage{
		page: page{
			Links:       getNavbarLinks(req),
			Admin:       Admin(req),
			Shopify:     shopifyKey,
			Name:        name,
			Stylesheets: []string{"/css/product.css"},
		},
		Product:     *p,
		Back:        back,
		BackText:    backText,
		Subcategory: subcat,
	}
	return templates.Get("product.html").ExecuteTemplate(w, "base", page)
}
