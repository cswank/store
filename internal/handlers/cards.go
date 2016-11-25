package handlers

import "net/http"

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

}
