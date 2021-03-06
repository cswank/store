package templates

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
)

var (
	templates map[string]tmpl

	multiplexer = template.FuncMap{
		"active": func(item, page string) string {
			if item == page {
				return "active-link"
			}
			return "inactive-link"
		},
		"printDate": func(date string) string {
			parts := strings.Split(date, "-")
			return fmt.Sprintf("%s/%s/%s", parts[1], parts[2], parts[0])
		},
		"getDate": func(ts time.Time) string {
			return ts.Format("01/02/2006")
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
)

type tmpl struct {
	template *template.Template
	files    []string
	funcs    template.FuncMap
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
		"about.html":                      {files: []string{"about.html"}},
		"admin/admin.html":                {files: []string{"admin/admin.html", "admin/links.html", "background-images.html", "admin/admin.js"}},
		"admin/category.html":             {files: []string{"admin/category.html", "admin/links.html", "background-images.html", "admin/admin.js"}},
		"admin/subcategory.html":          {files: []string{"admin/subcategory.html", "admin/links.html", "background-images.html", "admin/admin.js"}},
		"admin/blogs.html":                {files: []string{"admin/blogs.html"}},
		"admin/product.html":              {files: []string{"admin/product.html", "admin/links.html", "admin/product.js", "background-images.html"}},
		"admin/wholesaler.html":           {files: []string{"admin/wholesaler.html"}},
		"admin/wholesalers.html":          {files: []string{"admin/wholesalers.html"}},
		"blogs/blogs.html":                {files: []string{"blogs/blogs.html"}, funcs: multiplexer},
		"admin/blog-form.html":            {files: []string{"admin/blog-form.html", "admin/blog.js"}, funcs: multiplexer},
		"cart.html":                       {files: []string{"cart.html", "cart.js"}},
		"category.html":                   {files: []string{"category.html", "thumb.html"}},
		"confirm.html":                    {files: []string{"confirm.html", "confirm.js"}},
		"contact.html":                    {files: []string{"contact.html"}},
		"index.html":                      {files: []string{"index.html"}},
		"lineitem.html":                   {files: []string{"lineitem.html"}, bare: true},
		"login.html":                      {files: []string{"login.html"}},
		"logout.html":                     {files: []string{"logout.html", "confirm.js"}},
		"notfound.html":                   {files: []string{"notfound.html"}},
		"product.html":                    {files: []string{"product.html", "product.js"}},
		"reset-form.html":                 {files: []string{"reset-form.html"}},
		"reset.html":                      {files: []string{"reset.html"}},
		"shop.html":                       {files: []string{"shop.html", "thumb.html"}, funcs: multiplexer},
		"subcategory.html":                {files: []string{"subcategory.html", "thumb.html"}, funcs: multiplexer},
		"wholesale/application-form.html": {files: []string{"wholesale/application-form.html", "wholesale/application.js"}},
		"wholesale/form.html":             {files: []string{"wholesale/form.html", "wholesale/thumb.html", "wholesale/wholesale.js"}, funcs: multiplexer},
		"wholesale/invoice-sent.html":     {files: []string{"wholesale/invoice-sent.html"}},
		"wholesale/invoice.html":          {files: []string{"wholesale/invoice.html"}},
		"wholesale/pending.html":          {files: []string{"wholesale/pending.html"}},
		"wholesale/preview.html":          {files: []string{"wholesale/preview.html"}},
		"wholesale/thanks.html":           {files: []string{"wholesale/thanks.html"}},
		"wholesale/welcome.html":          {files: []string{"wholesale/welcome.html"}},
	}

	base := []string{"head.html", "base.html", "navbar.html", "base.js"}

	for key, val := range templates {
		t := template.New(key)
		if val.funcs != nil {
			t = t.Funcs(val.funcs)
		}
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
