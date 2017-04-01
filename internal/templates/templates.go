package templates

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/GeertJohan/go.rice"
)

var (
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
	html := getHTML(box)
	for _, pth := range html {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(pth, err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"admin/product.html":              {files: []string{"admin/product.html", "admin/links.html", "admin/product.js", "background-images.html"}},
		"admin/wholesalers.html":          {files: []string{"admin/wholesalers.html"}},
		"admin/wholesaler.html":           {files: []string{"admin/wholesaler.html"}},
		"admin/admin.html":                {files: []string{"admin/admin.html", "admin/links.html", "background-images.html", "admin/admin.js"}},
		"cart.html":                       {files: []string{"cart.html", "cart.js"}},
		"category.html":                   {files: []string{"category.html", "thumb.html"}},
		"confirm.html":                    {files: []string{"confirm.html", "confirm.js"}},
		"contact.html":                    {files: []string{"contact.html"}},
		"index.html":                      {files: []string{"index.html"}},
		"lineitem.html":                   {files: []string{"lineitem.html"}, bare: true},
		"login.html":                      {files: []string{"login.html"}},
		"logout.html":                     {files: []string{"logout.html", "confirm.js"}},
		"notfound.html":                   {files: []string{"notfound.html"}},
		"product.html":                    {files: []string{"product.html", "shop.js"}},
		"reset.html":                      {files: []string{"reset.html"}},
		"reset-form.html":                 {files: []string{"reset-form.html"}},
		"shop.html":                       {files: []string{"shop.html", "thumb.html"}},
		"subcategory.html":                {files: []string{"subcategory.html", "thumb.html"}},
		"wholesale/application-form.html": {files: []string{"wholesale/application-form.html", "wholesale/application.js"}},
		"wholesale/preview.html":          {files: []string{"wholesale/preview.html"}},
		"wholesale/thanks.html":           {files: []string{"wholesale/thanks.html"}},
		"wholesale/form.html":             {files: []string{"wholesale/form.html", "wholesale/thumb.html", "quantity.js"}},
		"wholesale/invoice.html":          {files: []string{"wholesale/invoice.html"}},
		"wholesale/invoice-sent.html":     {files: []string{"wholesale/invoice-sent.html"}},
		"wholesale/pending.html":          {files: []string{"wholesale/pending.html"}},
		"wholesale/welcome.html":          {files: []string{"wholesale/welcome.html"}},
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

func getHTML(box *rice.Box) []string {
	fmt.Println("getHTML")
	var html []string
	box.Walk("/", func(pth string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(pth, ".html") || strings.HasSuffix(pth, ".js") {
			if box.IsEmbedded() {
				pth = pth[1:] //workaround until https://github.com/GeertJohan/go.rice/issues/71 is fixed (which is probably never)
			}
			html = append(html, pth)
		}
		return nil
	})
	return html
}

func Get(k string) *template.Template {
	return templates[k].template
}
