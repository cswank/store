package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/storage"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/schema"
)

type loginPage struct {
	templates.Page
	Resource       string
	Captcha        bool
	CaptchaSiteKey string
}

func Login(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		Page: templates.Page{
			Links:   templates.GetLinks(),
			Scripts: []string{"https://www.google.com/recaptcha/api.js"},
		},
		Captcha: true,
	}

	return templates.Get("login.html").ExecuteTemplate(w, "base", p)
}

func DoLogin(w http.ResponseWriter, req *http.Request) error {

	err := req.ParseForm()
	if err != nil {
		return err
	}

	var u storage.User
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
		Page: templates.Page{
			Links: templates.GetLinks(),
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
