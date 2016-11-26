package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
)

var (
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
	for _, pth := range []string{"head.html", "base.html", "navbar.html", "index.html", "login.html", "item-form.html", "items.html", "item.html", "thumb.html", "subcategory.html", "admin.html", "category-admin.html", "category-admin.js", "confirm.html", "confirm.js"} {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"index.html":          {files: []string{"index.html"}},
		"login.html":          {files: []string{"login.html"}},
		"item-form.html":      {files: []string{"item-form.html"}},
		"items.html":          {files: []string{"items.html"}},
		"item.html":           {files: []string{"item.html"}},
		"subcategory.html":    {files: []string{"subcategory.html", "thumb.html"}},
		"admin.html":          {files: []string{"admin.html"}},
		"confirm.html":        {files: []string{"confirm.html", "confirm.js"}},
		"category-admin.html": {files: []string{"category-admin.html", "category-admin.js"}},
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
	fmt.Println("ico")
	w.Write(ico)
}
