package handlers

import (
	"log"
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/schema"
)

type loginPage struct {
	page
}

func Login(w http.ResponseWriter, req *http.Request) {
	var p loginPage
	templates["login.html"].template.ExecuteTemplate(w, "base", p)
}

func DoLogin(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Println("err", err)
		return
	}

	var u store.User
	if err := schema.NewDecoder().Decode(&u, req.PostForm); err != nil {
		log.Println("err", err)
		return
	}

	ok, err := u.CheckPassword()
	if !ok || err != nil {
		lg.Println("bad request")
		w.Header().Set("Location", "/login.html")
		w.WriteHeader(http.StatusFound)
		return
	}
	http.SetCookie(w, getCookie(u.Email))
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func Logout(w http.ResponseWriter, req *http.Request) {
	cookie := &http.Cookie{
		Name:   "quimby",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Location", "/login.html")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
