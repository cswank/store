package handlers

import (
	"log"
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/schema"
)

type cardPage struct {
	page
}

func Cards(w http.ResponseWriter, req *http.Request) {
	p := cardPage{}
	templates["index.html"].template.ExecuteTemplate(w, "base", p)
}

func CardForm(w http.ResponseWriter, req *http.Request) {
	p := cardPage{}
	templates["card-form.html"].template.ExecuteTemplate(w, "base", p)
}

func CardFormUpdate(w http.ResponseWriter, req *http.Request) {
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
