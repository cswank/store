package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/cswank/store/internal/config"
)

var (
	cfg          config.Config
	mailTemplate = `From: %s
To: %s
Subject: %s

%s
`
	Send func(Msg) error
)

type Msg struct {
	Name    string `schema:"name"`
	From    string `schema:"from"`
	To      string `schema:"to"`
	Subject string `schema:"subject"`
	Body    string `schema:"body"`
}

func Init(c config.Config) {
	cfg = c
	if cfg.Email == "" || cfg.EmailPassword == "" {
		log.Println("warning: STORE_EMAIL or STORE_EMAIL_PASSWORD not set, using fake email (writes to /tmp/mail.txt)")
		Send = sendFake
	} else {
		Send = sendEmail
	}
}

func sendEmail(msg Msg) error {
	text := fmt.Sprintf(mailTemplate, msg.From, msg.To, msg.Subject, msg.Body)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", cfg.Email, cfg.EmailPassword, "smtp.gmail.com"),
		cfg.Email,
		[]string{msg.To}, []byte(text),
	)
}

func sendFake(msg Msg) error {
	text := fmt.Sprintf(mailTemplate, msg.From, msg.To, msg.Subject, msg.Body)

	f, err := os.Create("/tmp/mail.txt")
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(text)
	return err
}
