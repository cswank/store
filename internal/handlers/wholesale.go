package handlers

import "net/http"

func Wholesale(w http.ResponseWriter, req *http.Request) {
	p := page{
		Links: getNavbarLinks(),
	}
	templates["wholesale.html"].template.ExecuteTemplate(w, "base", p)
}
