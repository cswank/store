package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type itemsPage struct {
	page
	Categories map[string][]string
}

type itemPage struct {
	page
	Item store.Item
}

func Items(w http.ResponseWriter, req *http.Request) {
	cat, err := store.GetCategories()
	if err != nil {
		log.Println("err", err)
		return
	}

	p := itemsPage{
		Categories: cat,
	}
	fmt.Println(p)
	templates["items.html"].template.ExecuteTemplate(w, "base", p)
}

type categoryPage struct {
	page
	Items []store.Item
}

func SubCategory(w http.ResponseWriter, req *http.Request) {
	args := req.URL.Query()
	vars := mux.Vars(req)
	fmt.Println("vars", vars, req.URL.Path)
	var p categoryPage
	var page int64
	pg := args.Get("page")
	if pg == "" {
		page = 1
	} else {
		var err error
		page, err = strconv.ParseInt(pg, 64, 10)
		if err != nil {
			log.Println("err", err)
			return
		}
	}

	items, err := store.GetSubCatetory(vars["category"], vars["subcategory"], int(page))
	if err != nil {
		log.Println("err", err)
		return
	}
	p.Items = items

	fmt.Println(p)
	templates["subcategory.html"].template.ExecuteTemplate(w, "base", p)
}

func ItemForm(w http.ResponseWriter, req *http.Request) {
	args := req.URL.Query()
	var p itemPage
	id := args.Get("id")
	if id != "" {
		item := store.Item{ID: id}
		if err := item.Fetch(); err != nil {
			log.Println("err", err)
			return
		}
		p.Item = item
	}
	templates["item-form.html"].template.ExecuteTemplate(w, "base", p)
}

func ItemFormUpdate(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Println("err", err)
		return
	}

	i, err := store.NewItem()
	if err != nil {
		log.Println("err", err)
		return
	}

	if err := schema.NewDecoder().Decode(i, req.PostForm); err != nil {
		log.Println("err", err)
		return
	}

	if err := i.Save(); err != nil {
		log.Println("err", err)
		return
	}

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
