package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"

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
			Admin: Admin(getUser(req)),
			Name:  name,
		},
	}

	if captcha {
		p.Captcha = true
		p.Scripts = []string{"https://www.google.com/recaptcha/api.js"}
		p.CaptchaSiteKey = captchaSiteKey
	}

	if s == "true" {
		p.ShowMessage = true
	}

	return templates["contact.html"].template.ExecuteTemplate(w, "base", p)
}

type msg struct {
	Name            string `schema:"name"`
	Email           string `schema:"email"`
	Subject         string `schema:"subject"`
	Body            string `schema:"body"`
	CaptchaResponse string `schema:"g-recaptcha-response"`
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
	if err := schema.NewDecoder().Decode(&m, req.PostForm); err != nil {
		return err
	}

	form := url.Values{}
	form.Add("secret", captchaSecretKey)
	form.Add("response", m.CaptchaResponse)
	resp, err := http.Post(captchaURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var c captchaResp
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return err
	}

	if c.Success {
	} else {
		lg.Println("invalid captcha", c.Errors)
		return nil
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
