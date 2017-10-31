package handlers

import (
	"fmt"
	"net/http"

	"github.com/cswank/store/internal/email"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/schema"
)

var (
	captcha bool
)

type contactPage struct {
	page

	Captcha        bool
	CaptchaSiteKey string
	ShowMessage    bool
}

func Contact(w http.ResponseWriter, req *http.Request) error {
	s := req.URL.Query().Get("submitted")
	p := contactPage{
		page: page{
			Links: getNavbarLinks(req),
			Admin: Admin(req),
			Name:  name,
			Head:  html["head"],
		},
	}

	if captcha {
		p.Captcha = true
		p.CaptchaSiteKey = cfg.RecaptchaSiteKey
	}

	if s == "true" {
		p.ShowMessage = true
	}

	return templates.Get("contact.html").ExecuteTemplate(w, "base", p)
}

type captchaResp struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

func DoContact(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	var m email.Msg
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&m, req.PostForm); err != nil {
		return err
	}

	m.Body = fmt.Sprintf("New message from %s for %s:\n\n%s", m.From, cfg.Domains[0], m.Body)
	m.To = cfg.Email

	if err := email.Send(m); err != nil {
		return err
	}

	w.Header().Set("Location", "/contact?submitted=true")
	w.WriteHeader(http.StatusFound)
	return nil
}
