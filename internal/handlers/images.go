package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

var (
	etags map[string]string
	eLock sync.Mutex
)

func Image(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	img, err := store.GetImage(vars["type"], vars["title"], vars["size"])
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
	return nil
}

func SiteImage(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	img, err := store.GetSiteImage(vars["title"])
	if err != nil {
		return err
	}

	setEtag(w, vars["title"], img)

	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
	return nil
}

func setEtag(w http.ResponseWriter, title string, img []byte) {
	t := fmt.Sprintf("%x", md5.Sum(img))
	eLock.Lock()
	etags[title] = t
	eLock.Unlock()
	w.Header().Set("Etag", t)
}

//ETag short-circuts the request if the client already has this resource.
func ETag(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if getMatch(req) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		h.ServeHTTP(w, req)
	})
}

func getMatch(req *http.Request) bool {
	vars := mux.Vars(req)

	eLock.Lock()
	t, ok := etags[vars["title"]]
	eLock.Unlock()
	if !ok {
		return false
	}

	match := req.Header.Get("If-None-Match")
	return match != "" && strings.Contains(match, t)
}
