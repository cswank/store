package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
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
	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
	return nil
}
