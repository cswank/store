package handlers

import (
	"net/http"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

func DeleteCategory(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	cat := store.Category{ID: id}
	if err := cat.Delete(); err != nil {
		lg.Println("couldn't delete category", id, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
