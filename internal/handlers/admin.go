package handlers

import (
	"log"
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/schema"
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
	q := req.URL.Query()
	p := categoryAdminPage{
		Name: q.Get("category"),
	}
	templates["category-admin.html"].template.ExecuteTemplate(w, "base", p)
}

func CategoryAdminFormHandler(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Println("err", err)
		return
	}

	var c store.Category
	if err := schema.NewDecoder().Decode(&c, req.PostForm); err != nil {
		log.Println("err", err)
		return
	}

	c.Save()
	w.Header().Set("Location", "/admin")
	w.WriteHeader(http.StatusFound)
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
