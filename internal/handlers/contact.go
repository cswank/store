package handlers

import (
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/schema"
)

var (
	storeEmail         string
	storeEmailPassword string
	captchaSiteKey     string
	captchaSecretKey   string
	captchaURL         string
	captcha            bool
	mailTemplate       = `From: %s
To: %s
Subject: %s

%s
`
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
		},
	}

	if captcha {
		p.Captcha = true
		p.Scripts = []string{"https://www.google.com/recaptcha/api.js"}
		p.CaptchaSiteKey = cfg.RecaptchaSiteKey
	}

	if s == "true" {
		p.ShowMessage = true
	}

	return templates.Get("contact.html").ExecuteTemplate(w, "base", p)
}

type msg struct {
	Name    string `schema:"name"`
	Email   string `schema:"email"`
	Subject string `schema:"subject"`
	Body    string `schema:"body"`
}

type captchaResp struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

func DoContact(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	var m msg
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&m, req.PostForm); err != nil {
		return err
	}

	if err := sendEmail(m); err != nil {
		return err
	}

	w.Header().Set("Location", "/contact?submitted=true")
	w.WriteHeader(http.StatusFound)
	return nil
}

func sendEmail(m msg) error {
	msg := fmt.Sprintf(mailTemplate, m.Email, storeEmail, m.Subject, m.Body)

	return smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", storeEmail, storeEmailPassword, "smtp.gmail.com"),
		m.Email,
		[]string{storeEmail}, []byte(msg),
	)
}
