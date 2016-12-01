package handlers

import "net/http"

type itemsPage struct {
	page
	Categories map[string][]string
}

type itemPage struct {
	page
	Categories    []string
	SubCategories []string
}

func Items(w http.ResponseWriter, req *http.Request) {
	p := itemsPage{}
	templates["items.html"].template.ExecuteTemplate(w, "base", p)
}

type item struct {
	Admin bool
	Edit  string
}

type categoryPage struct {
	page
	Admin bool
}

func SubCategory(w http.ResponseWriter, req *http.Request) {
	admin := Admin(getUser(req))
	p := categoryPage{
		Admin: admin,
	}

	templates["subcategory.html"].template.ExecuteTemplate(w, "base", p)
}
