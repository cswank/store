package templates

import (
	"html/template"
	"log"

	"github.com/GeertJohan/go.rice"
)

var (
	html = []string{
		"base.html",
		"base.js",
		"cart.html",
		"cart.js",
		"contact.html",
		"head.html",
		"index.html",
		"lineitem.html",
		"navbar.html",
		"product.html",
		"shop.js",
		"subcategory.html",
		"thumb.html",
	}

	templates map[string]tmpl
)

type tmpl struct {
	template *template.Template
	files    []string
	bare     bool
}

func Get(name string) *template.Template {
	return templates[name].template
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
		"cart.html":        {files: []string{"cart.html", "cart.js"}},
		"contact.html":     {files: []string{"contact.html"}},
		"index.html":       {files: []string{"index.html"}},
		"lineitem.html":    {files: []string{"lineitem.html"}, bare: true},
		"product.html":     {files: []string{"product.html", "shop.js"}},
		"subcategory.html": {files: []string{"subcategory.html", "thumb.html"}},
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
