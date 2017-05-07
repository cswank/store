package handlers

import (
	"html/template"
	"net/http"

	"github.com/cswank/store/internal/templates"
)

type aboutPage struct {
	page
	Body template.HTML
}

func About(w http.ResponseWriter, req *http.Request) error {
	p := aboutPage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		Body: html["about"],
	}

	return templates.Get("about.html").ExecuteTemplate(w, "base", p)
}
