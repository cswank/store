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

type item struct {
	store.Item
	Admin bool
	Edit  string
}

type categoryPage struct {
	page
	Items []item
	Admin bool
}

func SubCategory(w http.ResponseWriter, req *http.Request) {
	args := req.URL.Query()
	vars := mux.Vars(req)
	admin := Admin(getUser(req))

	p := categoryPage{
		Admin: admin,
	}
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
	cat := vars["category"]
	subcat := vars["subcategory"]
	it, err := store.GetSubCatetory(cat, subcat, int(page))
	if err != nil {
		log.Println("err", err)
		return
	}
	items := make([]item, len(it))
	for i, x := range it {
		items[i] = item{Item: x, Admin: admin, Edit: fmt.Sprintf("/admin/items/edit?id=%s&category=%s&subcategory=%s&page=%d", x.ID, cat, subcat, page)}
	}
	p.Items = items
	templates["subcategory.html"].template.ExecuteTemplate(w, "base", p)
}

func ItemForm(w http.ResponseWriter, req *http.Request) {
	args := req.URL.Query()
	var p itemPage
	id := args.Get("id")
	s := args.Get("page")
	page, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Println("err", err)
		return
	}

	if id != "" {
		item := store.Item{ID: id, Category: args.Get("category"), SubCategory: args.Get("subcategory"), Page: int(page)}
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

	fmt.Println("done saving item")
	w.Header().Set("Location", "/admin")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
