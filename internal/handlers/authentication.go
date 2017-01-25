package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/securecookie"
)

var (
	hashKey, blockKey      []byte
	sc                     *securecookie.SecureCookie
	domain, authCookieName string
)

func Authentication(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user, err := getUserFromCookie(req)
		if err != nil && err != http.ErrNoCookie {
			ctx := context.WithValue(req.Context(), "error", err)
			req = req.WithContext(ctx)
		} else {
			ctx := context.WithValue(req.Context(), "user", user)
			req = req.WithContext(ctx)
		}
		h.ServeHTTP(w, req)
	})
}

func getCookie(email string) *http.Cookie {
	val := map[string]string{
		"email": email,
	}

	encoded, err := sc.Encode(authCookieName, val)
	fmt.Println(encoded, authCookieName)
	if err != nil {
		log.Println("couldn't encode cookie", err)
	}
	return &http.Cookie{
		Name:     authCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	}
}

func getUserFromCookie(req *http.Request) (*store.User, error) {
	user := &store.User{}
	cookie, err := req.Cookie(authCookieName)
	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = sc.Decode(authCookieName, cookie.Value, &m)
	if err != nil {
		return nil, err
	}

	if m["email"] == "" {
		return nil, errors.New("no way, eh")
	}
	user.Email = m["email"]
	err = user.Fetch()
	user.HashedPassword = []byte{}
	return user, err
}
