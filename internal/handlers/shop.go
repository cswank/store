package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/cswank/store/internal/store"
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
	return templates["shop.html"].template.ExecuteTemplate(w, "base", p)
}

type cartPage struct {
	page
	Price string
}

func Cart(w http.ResponseWriter, req *http.Request) error {

	p := cartPage{
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Shopify: shopify,
			Name:    name,
		},
		Price: store.DefaultPrice,
	}
	return templates["cart.html"].template.ExecuteTemplate(w, "base", p)
}

func LineItem(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	p := store.NewProduct(vars["title"], vars["category"], vars["subcategory"], "")
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
	price, err := strconv.ParseFloat(store.DefaultPrice, 10)
	if err != nil {
		return err
	}

	t := float64(q) * price
	p.Total = fmt.Sprintf("%.02f", t)
	return templates["lineitem.html"].template.ExecuteTemplate(w, "lineitem.html", p)
}

type categoryPage struct {
	page
	SubCategories []link
}

func Category(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	subs, err := store.GetSubCategories(vars["category"])
	if err != nil {
		return err
	}

	p := categoryPage{
		SubCategories: getLinks(fmt.Sprintf("/shop/%s", vars["category"]), subs),
		page: page{
			Admin:   Admin(req),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
	}
	return templates["category.html"].template.ExecuteTemplate(w, "base", p)
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
			Image: fmt.Sprintf("/images/products/%s/thumb.png", t),
			Link:  fmt.Sprintf("/shop/%s/%s/%s", cat, subcat, t),
			Price: store.DefaultPrice,
		}
	}
	return out
}

func SubCategory(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	prods, err := store.GetProducts(vars["category"], vars["subcategory"])
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
		Products: getProducts(vars["category"], vars["subcategory"], prods),
	}
	return templates["subcategory.html"].template.ExecuteTemplate(w, "base", p)
}

type productPage struct {
	page
	Product store.Product
}

type product struct {
	Title     string
	Image     string
	Link      string
	ProductID string
	Price     string
}

func GetProduct(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	p := store.NewProduct(vars["title"], vars["category"], vars["subcategory"], "")
	if err := p.Fetch(); err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(p)
}

func Product(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	p := store.NewProduct(vars["title"], vars["category"], vars["subcategory"], "")

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
			Stylesheets: []string{"/static/css/product.css"},
		},
		Product: *p,
	}
	return templates["product.html"].template.ExecuteTemplate(w, "base", page)
}
