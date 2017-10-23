package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cswank/store/internal/email"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/schema"
)

type loginPage struct {
	page
	Resource       string
	Captcha        bool
	CaptchaSiteKey string
	Error          string
	Message        string
	Token          string
	Email          string
	Action         string
}

func Login(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		page: page{
			Links: getNavbarLinks(req),
			Admin: Admin(req),
			Head:  html["head"],
		},
		Captcha:        true,
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
		Error:          req.URL.Query().Get("error"),
	}

	return templates.Get("login.html").ExecuteTemplate(w, "base", p)
}

func DoLogin(w http.ResponseWriter, req *http.Request) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}

	var u store.User
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&u, req.PostForm); err != nil {
		return err
	}

	ok, err := u.CheckPassword()
	if !ok || err != nil {
		return errInvalidLogin
	}

	http.SetCookie(w, getCookie(u.Email))
	if isAdmin(&u) {
		w.Header().Set("Location", "/admin")
	} else {
		w.Header().Set("Location", "/wholesale")
	}
	w.WriteHeader(http.StatusFound)
	return nil
}

func Logout(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		page: page{
			Links: getNavbarLinks(req),
			Admin: Admin(req),
			Head:  html["head"],
		},
	}
	return templates.Get("logout.html").ExecuteTemplate(w, "base", p)
}

func DoLogout(w http.ResponseWriter, req *http.Request) error {
	cookie := &http.Cookie{
		Name:   authCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
	return nil
}

func ResetPage(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		page: page{
			Links: getNavbarLinks(req),
			Admin: Admin(req),
			Head:  html["head"],
		},
		Message:        req.URL.Query().Get("message"),
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
	}
	return templates.Get("reset.html").ExecuteTemplate(w, "base", p)
}

func SendReset(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	em := req.FormValue("email")

	token, err := store.SavePasswordReset(em)
	if err != nil {
		return err
	}

	if err := sendPasswordResetEmail(em, token); err != nil {
		return err
	}

	msg := "You will soon receive an email that contains a link to reset your password."
	w.Header().Set("Location", fmt.Sprintf("/login/reset?message=%s", msg))
	w.WriteHeader(http.StatusFound)
	return nil
}

func sendPasswordResetEmail(em, token string) error {
	d := cfg.Domains[0]
	if cfg.Port != 443 && cfg.Port != 80 {
		d = fmt.Sprintf("%s:%d", d, cfg.Port)
	}
	u, _ := url.Parse(fmt.Sprintf("https://%s/login/do-reset", d))
	params := url.Values{
		"token":    {token},
		"username": {em},
	}
	u.RawQuery = params.Encode()

	body := `Dear %s,
Please click on the following link to reset your %s password.

%s

Sincerely,
%s
`
	m := email.Msg{
		Email:   em,
		Subject: fmt.Sprintf("%s password reset request", cfg.Name),
		Body:    fmt.Sprintf(body, em, cfg.Domains[0], u, cfg.Domains[0]),
	}

	return email.Send(m)
}

func ResetPassword(w http.ResponseWriter, req *http.Request) error {
	t := req.URL.Query().Get("token")
	if t == "" {
		return store.ErrNotFound
	}

	u, err := store.GetUserFromResetToken(t, false)
	if err != nil {
		return err
	}

	p := loginPage{
		page: page{
			Links: getNavbarLinks(req),
			Admin: Admin(req),
			Head:  html["head"],
		},
		Message:        req.URL.Query().Get("message"),
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
		Token:          t,
		Email:          u.Email,
		Action:         fmt.Sprintf("/login/do-reset?username=%s&token=%s", u.Email, t),
	}

	return templates.Get("reset-form.html").ExecuteTemplate(w, "base", p)
}

func DoResetPassword(w http.ResponseWriter, req *http.Request) error {
	t := req.URL.Query().Get("token")
	if t == "" {
		return store.ErrNotFound
	}

	if err := req.ParseForm(); err != nil {
		return err
	}

	pw := req.FormValue("password")
	pw2 := req.FormValue("confirm-password")

	u, err := store.GetUserFromResetToken(t, true)
	if err != nil {
		return err
	}

	u.Password = pw
	u.Password2 = pw2

	if err := u.UpdatePassword(); err != nil {
		return err
	}

	w.Header().Set("Location", "/wholesale")
	w.WriteHeader(http.StatusFound)
	return nil
}
