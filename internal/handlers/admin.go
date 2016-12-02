package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cswank/store/internal/store"
)

func AdminPage(w http.ResponseWriter, req *http.Request) {
	p := page{}
	templates["admin.html"].template.ExecuteTemplate(w, "base", p)
}

func AddItems(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		log.Println("err", err)
		return
	}

	ff, _, err := req.FormFile("Items")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ff.Close()

	if err := store.ImportItems(ff); err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Location", "/")
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
