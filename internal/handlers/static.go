package handlers

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
)

var (
	html = []string{
		"head.html",
		"base.html",
		"navbar.html",
		"index.html",
		"login.html",
		"contact.html",
		"wholesale.html",
		"shop.html",
		"category.html",
		"subcategory.html",
		"item.html",
		"thumb.html",
		"admin.html",
		"confirm.html",
		"confirm.js",
		"shopify.html",
	}

	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
	templates              map[string]tmpl
	ico                    []byte
)

type tmpl struct {
	template *template.Template
	files    []string
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
		"index.html":       {files: []string{"index.html"}},
		"login.html":       {files: []string{"login.html"}},
		"contact.html":     {files: []string{"contact.html"}},
		"wholesale.html":   {files: []string{"wholesale.html"}},
		"shop.html":        {files: []string{"shop.html", "thumb.html"}},
		"category.html":    {files: []string{"category.html", "thumb.html"}},
		"subcategory.html": {files: []string{"subcategory.html", "thumb.html"}},
		"item.html":        {files: []string{"item.html", "shopify.html"}},
		"admin.html":       {files: []string{"admin.html"}},
		"confirm.html":     {files: []string{"confirm.html", "confirm.js"}},
	}

	base := []string{"head.html", "base.html", "navbar.html"}

	for key, val := range templates {
		t := template.New(key)
		var err error
		for _, f := range append(val.files, base...) {
			t, err = t.Parse(data[f])
			if err != nil {
				log.Fatal(err)
			}
		}
		val.template = t
		templates[key] = val
	}

	f, err := box.Open("favicon.ico")

	if err == nil {
		ico, _ = ioutil.ReadAll(f)
		f.Close()
	}

}

func Favicon(w http.ResponseWriter, req *http.Request) {

}
