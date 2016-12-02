package handlers

import "net/http"

type page struct {
	Admin bool
}

func Home(w http.ResponseWriter, req *http.Request) {
	p := page{}
	templates["index.html"].template.ExecuteTemplate(w, "base", p)
}
