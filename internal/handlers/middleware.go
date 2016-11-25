package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/cswank/store/internal/store"
	"github.com/justinas/alice"
)

var (
	logFile *os.File
	lg      *log.Logger
)

func getLogger(logOutput string) *log.Logger {
	if logOutput == "stdout" {
		return log.New(os.Stdout, "store", log.Lshortfile)
	}

	var err error
	if logFile, err = os.Create("/tmp/store.log"); err != nil {
		log.Fatal(err)
	}
	return log.New(logFile, "store", log.Lshortfile)
}

func Close() {
	logFile.Close()
}

func Log(logOutput string) alice.Constructor {
	lg = getLogger(logOutput)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			lg.Println(req.URL.Path)
			h.ServeHTTP(w, req)
		})
	}
}

func Errors(w http.ResponseWriter, req *http.Request) {
	e := req.Context().Value("error")
	if e != nil {
		lg.Printf("error (%v)\n", e)
		err := e.(error)
		if err == store.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func Handle(f http.HandlerFunc) alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			e := req.Context().Value("error")
			if e == nil {
				f(w, req)
			}
			h.ServeHTTP(w, req)
		})
	}
}
