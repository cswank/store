package handlers

import "net/http"

func Home(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello world"))
}
