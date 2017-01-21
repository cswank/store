package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/justinas/alice"
)

var (
	logFile *os.File
	lg      *log.Logger
)

func Close() {
	logFile.Close()
}

func Log(logOutput string) alice.Constructor {
	lg = log.New(os.Stdout, "", 0)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			lg.Println(req.RemoteAddr, req.Method, req.URL.Path)
			h.ServeHTTP(w, req)
		})
	}
}
