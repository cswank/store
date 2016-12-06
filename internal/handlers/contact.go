package handlers

import "net/http"

func Contact(w http.ResponseWriter, req *http.Request) {
	p := page{
		Links: getNavbarLinks(),
	}
	templates["contact.html"].template.ExecuteTemplate(w, "base", p)
}
