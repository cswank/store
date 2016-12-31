package site

import (
	"html/template"
	"log"

	"github.com/GeertJohan/go.rice"
)

var (
	html = []string{
		"base.html",
		"base.js",
		"navbar.html",
		"product.html",
		"shop.js",
		"head.html",
	}

	templates map[string]tmpl
)

type tmpl struct {
	template *template.Template
	files    []string
	bare     bool
}

func Init(c Config, box *rice.Box) {
	cfg = c

	data := map[string]string{}
	for _, pth := range html {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"product.html": {files: []string{"product.html", "shop.js"}},
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
