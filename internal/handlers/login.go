package handlers

import (
	"net/http"

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
}

func Login(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		page: page{
			Links:   getNavbarLinks(req),
			Admin:   Admin(req),
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
		},
		Captcha:        true,
		CaptchaSiteKey: captchaSiteKey,
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
