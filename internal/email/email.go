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
	Email   string `schema:"email"`
	Subject string `schema:"subject"`
	Body    string `schema:"body"`
}

func Init(cfg config.Config) {
	fmt.Println(cfg.Email, cfg.EmailPassword)
	if cfg.Email == "" || cfg.EmailPassword == "" {
		log.Println("warning: STORE_EMAIL or STORE_EMAIL_PASSWORD not set, using fake email (writes to /tmp/mail.txt)")
		Send = sendFake
	} else {
		Send = sendEmail
	}
}

func sendEmail(m Msg) error {
	msg := fmt.Sprintf(mailTemplate, m.Email, cfg.Email, m.Subject, m.Body)

	return smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", cfg.Email, cfg.EmailPassword, "smtp.gmail.com"),
		m.Email,
		[]string{cfg.Email}, []byte(msg),
	)
}

func sendFake(msg Msg) error {
	text := fmt.Sprintf(mailTemplate, cfg.Email, "the developer", msg.Subject, msg.Body)

	f, err := os.OpenFile("/tmp/mail.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(text)
	return err
}
