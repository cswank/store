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

func getLogger(logOutput string) *log.Logger {
	f := log.Ldate | log.Ltime | log.Lmicroseconds
	if logOutput == "stdout" {
		return log.New(os.Stdout, "", f)
	}

	var err error
	if logFile, err = os.Create("/tmp/store.log"); err != nil {
		log.Fatal(err)
	}
	return log.New(logFile, "", f)
}

func Close() {
	logFile.Close()
}

func Log(logOutput string) alice.Constructor {
	lg = getLogger(logOutput)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			lg.Println(req.Method, req.URL.Path)
			h.ServeHTTP(w, req)
		})
	}
}
