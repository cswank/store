package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

type categoriesAdminPage struct {
	page
	Categories map[string][]string
}

func AdminPage(w http.ResponseWriter, req *http.Request) {
	cat, err := store.GetCategories()
	if err != nil {
		lg.Println(err)
		return
	}
	p := categoriesAdminPage{
		Categories: cat,
	}
	templates["admin.html"].template.ExecuteTemplate(w, "base", p)
}

type categoryAdminPage struct {
	page
	Name string
}

func CategoryAdmin(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["category"]
	p := categoryAdminPage{
		Name: name,
	}
	templates["category-admin.html"].template.ExecuteTemplate(w, "base", p)
}

type confirmPage struct {
	page
	Name     string
	Resource string
}

func Confirm(w http.ResponseWriter, req *http.Request) {
	args := req.URL.Query()
	name := args.Get("name")
	resource := args.Get("resource")
	p := confirmPage{
		Name:     name,
		Resource: resource,
	}
	templates["confirm.html"].template.ExecuteTemplate(w, "base", p)
}
