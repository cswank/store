package templates

import (
	"errors"
	"html/template"
	"log"

	"github.com/GeertJohan/go.rice"
)

var (
	html = []string{
		"admin-links.html",
		"admin-product.js",
		"admin-product.html",
		"admin.html",
		"admin.js",
		"background-images.html",
		"base.html",
		"base.js",
		"cart.html",
		"cart.js",
		"category.html",
		"confirm.html",
		"confirm.js",
		"contact.html",
		"index.html",
		"lineitem.html",
		"login.html",
		"logout.html",
		"navbar.html",
		"product.html",
		"shop.html",
		"shop.js",
		"subcategory.html",
		"thumb.html",
		"wholesale.html",
		"head.html",
	}

	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
	templates              map[string]tmpl
)

type tmpl struct {
	template *template.Template
	files    []string
	bare     bool
}

func Init(box *rice.Box) {
	data := map[string]string{}
	for _, pth := range html {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"admin-product.html": {files: []string{"admin-product.html", "admin-links.html", "admin-product.js", "background-images.html"}},
		"admin.html":         {files: []string{"admin.html", "admin-links.html", "background-images.html", "admin.js"}},
		"cart.html":          {files: []string{"cart.html", "cart.js"}},
		"category.html":      {files: []string{"category.html", "thumb.html"}},
		"confirm.html":       {files: []string{"confirm.html", "confirm.js"}},
		"contact.html":       {files: []string{"contact.html"}},
		"index.html":         {files: []string{"index.html"}},
		"lineitem.html":      {files: []string{"lineitem.html"}, bare: true},
		"login.html":         {files: []string{"login.html"}},
		"logout.html":        {files: []string{"logout.html", "confirm.js"}},
		"product.html":       {files: []string{"product.html", "shop.js"}},
		"shop.html":          {files: []string{"shop.html", "thumb.html"}},
		"subcategory.html":   {files: []string{"subcategory.html", "thumb.html"}},
		"wholesale.html":     {files: []string{"wholesale.html"}},
	}

	base := []string{"head.html", "base.html", "navbar.html", "base.js"}

	for key, val := range templates {
		t := template.New(key)
		var err error
		var items []string
		if val.bare {
			items = val.files
		} else {
			items = append(val.files, base...)
		}

		for _, f := range items {
			t, err = t.Parse(data[f])
			if err != nil {
				log.Fatal(err)
			}
		}
		val.template = t
		templates[key] = val
	}
}

func Get(k string) *template.Template {
	return templates[k].template
}
