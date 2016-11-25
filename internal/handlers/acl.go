package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/justinas/alice"
)

func getUser(req *http.Request) *store.User {
	u := req.Context().Value("user")
	if u == nil {
		return nil
	}
	return u.(*store.User)
}

type ACL func(user *store.User) bool

func Or(acls ...ACL) ACL {
	return func(user *store.User) bool {
		for _, f := range acls {
			if f(user) {
				return true
			}
		}
		return false
	}
}

func And(acls ...ACL) ACL {
	return func(user *store.User) bool {
		b := false
		for _, f := range acls {
			b = b && f(user)
		}
		return b
	}
}

func Admin(user *store.User) bool {
	return user != nil && user.Permission == store.Admin
}

func Read(user *store.User) bool {
	return user != nil && (user.Permission == store.Admin || user.Permission == store.Read)
}

func Anyone(user *store.User) bool {
	return true
}

func Perm(f ACL) alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user := getUser(req)
			if !f(user) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Not Authorized"))
				return
			}
			h.ServeHTTP(w, req)
		})
	}
}
