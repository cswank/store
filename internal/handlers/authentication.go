package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/securecookie"
)

var (
	hashKey  = []byte(os.Getenv("STORE_HASH_KEY"))
	blockKey = []byte(os.Getenv("STORE_BLOCK_KEY"))
	sc       = securecookie.New(hashKey, blockKey)
)

func getUserFromCookie(req *http.Request) (*store.User, error) {
	user := &store.User{}
	cookie, err := req.Cookie("store-user")
	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = sc.Decode("store-user", cookie.Value, &m)
	if err != nil {
		return nil, err
	}

	if m["user"] == "" {
		return nil, errors.New("no way, eh")
	}
	user.Username = m["user"]
	err = user.Fetch()
	user.HashedPassword = []byte{}
	user.TFAData = []byte{}
	return user, err
}

func ShortCircuit(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Context().Value("error") == nil {
			h.ServeHTTP(w, req)
		} else {
			e := req.Context().Value("error")
			if e != nil {
				log.Printf("error (%v)\n", e)
				err := e.(error)
				if err == store.ErrNotFound {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				w.Write([]byte(err.Error()))
			}
		}
	})
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
