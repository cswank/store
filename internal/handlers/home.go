package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/cswank/store/internal/templates"
)

type page struct {
	Shopify shopifyAPI
	Admin   bool
	Links   []link
	Name    string
	Message string
	Head    template.HTML
}

type homePage struct {
	page
	Home template.HTML
}

var (
	name = os.Getenv("STORE_NAME")
)

func Home(w http.ResponseWriter, req *http.Request) error {
	p := homePage{
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Shopify: shopifyKey,
			Name:    name,
			Head:    html["head"],
		},
		Home: html["home"],
	}

	return templates.Get("index.html").ExecuteTemplate(w, "base", p)
}

func Redirect(w http.ResponseWriter, req *http.Request) {
	http.Redirect(
		w,
		req,
		fmt.Sprintf("https://%s%s", req.Host, req.URL.String()),
		http.StatusMovedPermanently,
	)
}
