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
	for _, pth := range []string{"head.html", "base.html", "navbar.html", "index.html", "login.html", "card-form.html"} {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"index.html":     {files: []string{"index.html"}},
		"login.html":     {files: []string{"login.html"}},
		"card-form.html": {files: []string{"card-form.html"}},
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
