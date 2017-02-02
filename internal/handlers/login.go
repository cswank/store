package handlers

import (
	"fmt"
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/mux"
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
}

func Login(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
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
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
	return nil
}

func Logout(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		page: page{
			Links: getNavbarLinks(req),
			Admin: Admin(req),
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
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
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

	if err := store.SendPasswordReset(req.FormValue("email")); err != nil {
		return err
	}

	msg := "You will soon receive an email that contains a link to reset your password."
	w.Header().Set("Location", fmt.Sprintf("/login/reset?message=%s", msg))
	w.WriteHeader(http.StatusFound)
	return nil
}

func ResetPassword(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	t := vars["token"]

	//TODO
	// if !store.ValidToken(t) {

	// }

	p := loginPage{
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
		},
		Message:        req.URL.Query().Get("message"),
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
		Token:          t,
	}

	return templates.Get("reset-form.html").ExecuteTemplate(w, "base", p)
}

func DoResetPassword(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	t := req.FormValue("token")
	pw := req.FormValue("password")
	pw2 := req.FormValue("confirm-password")

	u, err := store.GetUserFromResetToken(t)
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
