package handlers

import (
	"fmt"
	"net/http"
)

type link struct {
	Name string
	Link string
}

type page struct {
	Admin bool
	Links []link
}

func Home(w http.ResponseWriter, req *http.Request) {
	p := page{
		Links: getNavbarLinks(),
	}
	templates["index.html"].template.ExecuteTemplate(w, "base", p)
}

func Redirect(w http.ResponseWriter, req *http.Request) {
	http.Redirect(
		w,
		req,
		fmt.Sprintf("https://%s%s", req.Host, req.URL.String()),
		http.StatusMovedPermanently,
	)
}
