package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/cswank/store/internal/store"
)

var (
	etags map[string]string
	eLock sync.Mutex
)

func Image(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars, err := getVars(req)
	if err != nil {
		return err
	}

	img, err := store.GetImage(vars["type"], vars["title"], vars["size"])
	if err != nil {
		return err
	}

	setEtag(w, req.URL.Path, img)
	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
	return nil
}

func setEtag(w http.ResponseWriter, pth string, img []byte) {
	t := fmt.Sprintf("%x", md5.Sum(img))
	eLock.Lock()
	etags[pth] = t
	eLock.Unlock()
	w.Header().Set("Etag", t)
}

//ETag short-circuts the request if the client already has this resource.
func ETag(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if matches(req) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		h.ServeHTTP(w, req)
	})
}

func matches(req *http.Request) bool {
	eLock.Lock()
	t, ok := etags[req.URL.Path]
	eLock.Unlock()
	if !ok {
		return false
	}

	match := req.Header.Get("If-None-Match")
	return match != "" && strings.Contains(match, t)
}
