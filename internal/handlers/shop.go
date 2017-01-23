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
	shopify shopifyAPI
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
		Shopify:    shopify,
		Categories: cats,
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Shopify: shopify,
		},
	}
	return templates.Get("shop.html").ExecuteTemplate(w, "base", p)
}

type cartPage struct {
	page
	Price             string
	UnderConstruction bool
}

func Cart(w http.ResponseWriter, req *http.Request) error {

	p := cartPage{
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Shopify: shopify,
			Name:    name,
		},
		Price:             cfg.DefaultPrice,
		UnderConstruction: cfg.UnderConstruction,
	}
	return templates.Get("cart.html").ExecuteTemplate(w, "base", p)
}

func LineItem(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat, "")
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
	price, err := strconv.ParseFloat(cfg.DefaultPrice, 10)
	if err != nil {
		return err
	}

	t := float64(q) * price
	p.Total = fmt.Sprintf("%.02f", t)
	return templates.Get("lineitem.html").ExecuteTemplate(w, "lineitem.html", p)
}

type categoryPage struct {
	page
	SubCategories []link
}

func Category(w http.ResponseWriter, req *http.Request) error {
	cat, _, _ := getVars(req)

	subs, err := store.GetSubCategories(cat)
	if err != nil {
		return err
	}

	p := categoryPage{
		SubCategories: getLinks(fmt.Sprintf("/shop/%s", cat), subs),
		page: page{
			Admin:   Admin(req),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
	}
	return templates.Get("category.html").ExecuteTemplate(w, "base", p)
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

func getProducts(cat, subcat string, prods []string) []product {
	out := make([]product, len(prods))
	for i, t := range prods {
		out[i] = product{
			Title: t,
			Image: fmt.Sprintf("/shop/images/products/%s/thumb.png", t),
			Link:  fmt.Sprintf("/shop/%s/%s/%s", cat, subcat, t),
			Price: cfg.DefaultPrice,
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

	prods, err := store.GetProducts(cat, subcat)
	if err != nil {
		return err
	}

	p := subCategoryPage{
		page: page{
			Admin:   Admin(req),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
		Products: getProducts(cat, subcat, prods),
	}
	return templates.Get("subcategory.html").ExecuteTemplate(w, "base", p)
}

type productPage struct {
	page
	Product     store.Product
	Back        string
	Subcategory string
}

type product struct {
	Title     string
	Image     string
	Link      string
	ProductID string
	Price     string
}

func GetProduct(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat, "")
	if err := p.Fetch(); err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(p)
}

func Product(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat, "")

	if err := p.Fetch(); err != nil {
		return err
	}

	p.Quantity = 1

	page := productPage{
		page: page{
			Links:       getNavbarLinks(req),
			Admin:       Admin(req),
			Shopify:     shopify,
			Name:        name,
			Stylesheets: []string{"/css/product.css"},
		},
		Product:     *p,
		Back:        fmt.Sprintf("/shop/%s/%s", cat, subcat),
		Subcategory: subcat,
	}
	return templates.Get("product.html").ExecuteTemplate(w, "base", page)
}
