package handlers

import "net/http"

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

func getNavbarLinks() []link {
	return []link{
		{Name: "Contact Us", Link: "/contact"},
		{Name: "Shop", Link: "/shop"},
	}
}
