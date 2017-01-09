package handlers

import (
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/gorilla/schema"
)

var (
	mailTemplate = `From: %s
To: %s
Subject: %s

%s
`
)

type msg struct {
	Name    string `schema:"name"`
	Email   string `schema:"email"`
	Subject string `schema:"subject"`
	Body    string `schema:"body"`
}

func Contact(w http.ResponseWriter, req *http.Request) error {
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
	msg := fmt.Sprintf(mailTemplate, m.Email, cfg.Email, m.Subject, m.Body)
	return smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", cfg.Email, cfg.EmailPassword, "smtp.gmail.com"),
		cfg.Email,
		[]string{cfg.Email}, []byte(msg),
	)
}
