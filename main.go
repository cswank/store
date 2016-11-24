package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/GeertJohan/go.rice"
	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/handlers"
	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var (
	cfg   store.Config
	serve = kingpin.Command("serve", "Start the server.")
	box   *rice.Box
)

const (
	version = "0.0.0"
)

func init() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	store.Init(cfg)
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("Craig Swank")
	switch kingpin.Parse() {
	case "serve":
		box = rice.MustFindBox("static")
		handlers.Init(box)
		Serve()
	}
}

func getMiddleware(perm handlers.ACL, f http.HandlerFunc) http.Handler {
	return alice.New(handlers.Authentication, handlers.Perm(perm), handlers.Handle(f)).Then(http.HandlerFunc(handlers.Errors))
}

func Serve() {
	r := mux.NewRouter()
	r.Handle("/", getMiddleware(handlers.Anyone, handlers.Home))
	r.Handle("/cards", getMiddleware(handlers.Anyone, handlers.Cards))

	r.Handle("/favicon.ico", getMiddleware(handlers.Anyone, handlers.Favicon))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(rice.MustFindBox("static").HTTPBox())))

	chain := alice.New(handlers.Log(cfg.LogOutput)).Then(r)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("listening on %s\n", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      chain,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
