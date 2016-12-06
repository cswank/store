package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/cswank/store/internal/shopify"
	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

type shopPage struct {
	page
	ShopifyDomain string
	ShopifyJSKey  string
	Categories    []string
}

func getItems(m map[string][]string) map[string][]item {
	out := map[string][]item{}
	for k, v := range m {
		items := make([]item, len(v))
		for i, n := range v {
			items[i] = item{
				Name:  n,
				Image: fmt.Sprintf("/items/Cards/%s/%s/thumb.jpg", k, n),
				Link:  fmt.Sprintf("/store/items/Cards/%s/%s", k, n),
			}
		}
		out[k] = items
	}
	return out
}

func getItem(cat, subcat, name string) (item, error) {
	m := store.GetCategory(cat)
	var out item
	v := m[subcat]

	for _, n := range v {
		if n == name {
			id, err := store.GetProductID(name)
			if err != nil {
				return out, err
			}

			out = item{
				Name:      n,
				Image:     fmt.Sprintf("/items/Cards/%s/%s/image.jpg", subcat, name),
				ProductID: id,
			}
			break
		}
	}
	return out, nil
}

func getSubcategory(cat, subcat string) []item {
	m := store.GetCategory(cat)
	var out []item
	v, ok := m[subcat]
	if !ok {
		return out
	}
	out = make([]item, len(v))
	for i, n := range v {
		out[i] = item{
			Name:  n,
			Image: fmt.Sprintf("/items/%s/%s/%s/thumb.jpg", cat, subcat, n),
			Link:  fmt.Sprintf("/shop/%s/%s/%s", cat, subcat, n),
		}
	}
	return out
}

func Shop(w http.ResponseWriter, req *http.Request) {
	p := shopPage{
		ShopifyDomain: shopify.ShopifyDomain,
		ShopifyJSKey:  shopify.ShopifyJSKey,
		Categories:    store.GetCategories(),
		page: page{
			Links: getNavbarLinks(),
		},
	}
	templates["shop.html"].template.ExecuteTemplate(w, "base", p)
}

type categoryPage struct {
	page
	SubCategories []link
}

func Category(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	p := categoryPage{
		SubCategories: getLinks(fmt.Sprintf("/shop/%s", vars["category"]), store.GetCategoryList(vars["category"])),
		page: page{
			Links: getNavbarLinks(),
		},
	}
	templates["category.html"].template.ExecuteTemplate(w, "base", p)
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
	Items []item
}

func SubCategory(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	p := subCategoryPage{
		page: page{
			Links: getNavbarLinks(),
		},
		Items: getSubcategory(vars["category"], vars["subcategory"]),
	}
	templates["subcategory.html"].template.ExecuteTemplate(w, "base", p)
}

type itemPage struct {
	page
	Item          item
	ShopifyDomain string
	ShopifyJSKey  string
}

type item struct {
	Name      string
	Image     string
	Link      string
	ProductID int
}

func Item(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	i, err := getItem(vars["category"], vars["subcategory"], vars["item"])
	if err != nil {
		log.Println("err", err)
		return
	}

	p := itemPage{
		page: page{
			Links: getNavbarLinks(),
			Admin: Admin(getUser(req)),
		},
		ShopifyDomain: shopify.ShopifyDomain,
		ShopifyJSKey:  shopify.ShopifyJSKey,
		Item:          i,
	}
	templates["item.html"].template.ExecuteTemplate(w, "base", p)
}
