package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/schema"
)

type formPage struct {
	page
	Captcha        bool
	CaptchaSiteKey string
	ShowMessage    bool
}

func Purchase(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func Wholesale(w http.ResponseWriter, req *http.Request) error {
	if Wholesaler(req) {
		return getWholesalePage(w, req)
	}
	return getWholesaleLogin(w, req)
}

type wholesalePage struct {
	page
	Products map[string]map[string][]product
}

func getWholesalePage(w http.ResponseWriter, req *http.Request) error {
	cats, err := store.GetCategories()
	if err != nil {
		return err
	}

	prods, err := getWholesaleProducts(cats)
	if err != nil {
		return err
	}

	p := wholesalePage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
		},
		Products: prods,
	}

	return templates.Get("wholesale-page.html").ExecuteTemplate(w, "base", p)
}

func getWholesaleProducts(cats []string) (map[string]map[string][]product, error) {
	m := map[string]map[string][]product{}
	for _, cat := range cats {
		subcats, err := store.GetSubCategories(cat)
		if err != nil {
			return nil, err
		}
		a := map[string][]product{}
		for _, subcat := range subcats {
			prods, err := store.GetProducts(cat, subcat)
			if err != nil {
				return nil, err
			}
			a[subcat] = getProducts(cat, subcat, prods)
		}
		m[cat] = a
	}
	return m, nil
}

func getWholesaleLogin(w http.ResponseWriter, req *http.Request) error {
	params := req.URL.Query()
	p := formPage{
		page: page{
			Links:   getNavbarLinks(req),
			Name:    cfg.Name,
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
		},
		ShowMessage:    params.Get("success") != "",
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
		Captcha:        true,
	}

	return templates.Get("wholesale-login.html").ExecuteTemplate(w, "base", p)
}

func WholesaleApply(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	u := store.User{
		Permission: store.Wholesaler,
	}

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&u, req.PostForm); err != nil {
		return err
	}

	if err := u.Save(); err != nil {
		return err
	}

	w.Header().Set("Location", "/wholesale?success=true")
	w.WriteHeader(http.StatusFound)
	return nil
}
