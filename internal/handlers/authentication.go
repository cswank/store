package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/securecookie"
)

var (
	hashKey, blockKey      []byte
	sc                     *securecookie.SecureCookie
	domain, authCookieName string
)

func init() {
	domain = os.Getenv("STORE_DOMAIN")
	if domain == "" {
		log.Fatal("you must set STORE_DOMAIN")
	}
	authCookieName = fmt.Sprintf("%s-user", domain)
	hashKey = []byte(os.Getenv("STORE_HASH_KEY"))
	blockKey = []byte(os.Getenv("STORE_BLOCK_KEY"))
	if string(hashKey) == "" || string(blockKey) == "" {
		log.Fatal("you must set STORE_HASH_KEY and STORE_BLOCK_KEY")
	}
	sc = securecookie.New(hashKey, blockKey)
}

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

	encoded, _ := sc.Encode(authCookieName, val)
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
	user.TFAData = []byte{}
	return user, err
}
