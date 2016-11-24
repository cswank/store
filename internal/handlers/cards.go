package handlers

import "net/http"

type cardPage struct {
	page
}

func Cards(w http.ResponseWriter, req *http.Request) {
	p := cardPage{}
	templates["index.html"].template.ExecuteTemplate(w, "base", p)
}
