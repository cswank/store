package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/schema"
)

type loginPage struct {
	page
	Resource string
}

func Login(w http.ResponseWriter, req *http.Request) error {
	p := loginPage{
		page: page{
			Links: getNavbarLinks(req),
			Admin: Admin(getUser(req)),
		},
	}
	return templates["login.html"].template.ExecuteTemplate(w, "base", p)
}

func DoLogin(w http.ResponseWriter, req *http.Request) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}

	var u store.User
	if err := schema.NewDecoder().Decode(&u, req.PostForm); err != nil {
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
			Admin: Admin(getUser(req)),
		},
	}
	return templates["logout.html"].template.ExecuteTemplate(w, "base", p)
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
