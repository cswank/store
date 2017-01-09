package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/storage"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/schema"
)

type formPage struct {
	templates.Page
	Captcha        bool
	CaptchaSiteKey string
	ShowMessage    bool
}

func Purchase(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func Wholesale(w http.ResponseWriter, req *http.Request) error {
	if Wholesaler(req) {
		return getWholesaleForm(w, req)

	}
	return getWholesalePage(w, req)
}

type wholesalePage struct {
	templates.Page
	Products map[string]map[string][]templates.Product
}

func getWholesalePage(w http.ResponseWriter, req *http.Request) error {
	cats := templates.GetCategories()
	p := wholesalePage{
		Page: templates.Page{
			Links: templates.GetLinks(),
			Name:  cfg.StoreName,
		},
		Products: getWholesaleProducts(cats),
	}

	return templates.Get("wholesale-page.html").ExecuteTemplate(w, "base", p)
}

func getWholesaleProducts(cats templates.Categories) map[string]map[string][]templates.Product {
	m := map[string]map[string][]templates.Product{}
	for cat, subcats := range cats {
		a := map[string][]templates.Product{}
		for subcat, prods := range subcats {
			a[subcat] = templates.GetProducts(cat, subcat, prods)
		}
		m[cat] = a
	}
	return m
}

func getWholesaleForm(w http.ResponseWriter, req *http.Request) error {
	params := req.URL.Query()
	p := formPage{
		Page: templates.Page{
			Links:   templates.GetLinks(),
			Name:    cfg.StoreName,
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
		},
		ShowMessage:    params.Get("success") != "",
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
		Captcha:        true,
	}

	return templates.Get("wholesale-form.html").ExecuteTemplate(w, "base", p)
}

func WholesalerRegistration(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	var ws storage.Wholesaler

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&ws, req.PostForm); err != nil {
		return err
	}

	if err := ws.Save(true); err != nil {
		return err
	}

	w.Header().Set("Location", "/wholesale?success=true")
	w.WriteHeader(http.StatusFound)
	return nil
}
