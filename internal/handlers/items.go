package handlers

import (
	"fmt"
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

type itemsPage struct {
	page
	Categories map[string][]item
}

type itemPage struct {
	page
	Item item
}

type item struct {
	Name  string
	Image string
	Link  string
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

func getItem(cat, subcat, name string) item {
	fmt.Println(cat, subcat, name)
	m := store.GetCategory(cat)
	var out item
	v := m[subcat]
	for _, n := range v {
		if n == name {
			out = item{
				Name:  n,
				Image: fmt.Sprintf("/items/Cards/%s/%s/image.jpg", subcat, name),
			}
		}
		break
	}
	return out
}

func Items(w http.ResponseWriter, req *http.Request) {
	p := itemsPage{
		Categories: getItems(store.GetCategory("Cards")),
	}
	templates["items.html"].template.ExecuteTemplate(w, "base", p)
}

type categoryPage struct {
	page
}

func SubCategory(w http.ResponseWriter, req *http.Request) {
	p := categoryPage{}
	templates["subcategory.html"].template.ExecuteTemplate(w, "base", p)
}

func Item(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	p := itemPage{
		page: page{
			Admin: Admin(getUser(req)),
		},
		Item: getItem(vars["category"], vars["subcategory"], vars["item"]),
	}
	templates["item.html"].template.ExecuteTemplate(w, "base", p)
}
