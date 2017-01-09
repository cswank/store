package templates

import (
	"html/template"
	"log"

	"github.com/GeertJohan/go.rice"
	"github.com/cswank/store/internal/config"
)

var (
	html = []string{
		"base.html",
		"base.js",
		"cart.html",
		"cart.js",
		"confirm.html",
		"confirm.js",
		"contact.html",
		"head.html",
		"index.html",
		"lineitem.html",
		"login.html",
		"logout.html",
		"navbar.html",
		"product.html",
		"quantity.js",
		"shop.js",
		"thumbs.html",
		"thumb.html",
		"wholesale-form.html",
		"wholesale-page.html",
		"wholesale-thumb.html",
	}

	templates map[string]tmpl
	cfg       config.Config
)

type tmpl struct {
	template *template.Template
	files    []string
	bare     bool
}

func Get(name string) *template.Template {
	return templates[name].template
}

func Init(c config.Config, box *rice.Box) {
	cfg = c
	initTemplates(box)
	initLinks()
}

func initTemplates(box *rice.Box) {
	data := map[string]string{}
	for _, pth := range html {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"cart.html":           {files: []string{"cart.html", "cart.js"}},
		"contact.html":        {files: []string{"contact.html"}},
		"index.html":          {files: []string{"index.html"}},
		"lineitem.html":       {files: []string{"lineitem.html"}, bare: true},
		"login.html":          {files: []string{"login.html"}},
		"logout.html":         {files: []string{"logout.html", "confirm.js"}},
		"product.html":        {files: []string{"product.html", "shop.js", "quantity.js"}},
		"thumbs.html":         {files: []string{"thumbs.html", "thumb.html"}},
		"wholesale-form.html": {files: []string{"wholesale-form.html"}},
		"wholesale-page.html": {files: []string{"wholesale-page.html", "wholesale-thumb.html", "quantity.js"}},
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
