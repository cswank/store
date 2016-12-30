package handlers

import "net/http"

func Wholesale(w http.ResponseWriter, req *http.Request) error {
	p := page{
		Links: getNavbarLinks(req),
		Admin: Admin(req),
		Name:  name,
	}
	return templates["wholesale.html"].template.ExecuteTemplate(w, "base", p)
}
