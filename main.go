package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var (
	cfg   config
	serve = kingpin.Command("serve", "Start the server.")
)

const (
	version = "0.0.0"
)

func init() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	Port      int    `env:"STORE_PORT" envDefault:"8080"`
	LogOutput string `env:"STORE_LOGOUTPUT" envDefault:"stdout"`
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("Craig Swank")
	switch kingpin.Parse() {
	case "serve":
		doServe()
	}
}

func getMiddleware(perm handlers.ACL, f http.HandlerFunc) http.Handler {
	return alice.New(handlers.Perm(perm)).Then(http.HandlerFunc(f))
}

func doServe() {
	r := mux.NewRouter()
	r.Handle("/", getMiddleware(handlers.Anyone, handlers.Home))

	chain := alice.New(handlers.Log(cfg.LogOutput), handlers.Authentication, handlers.ShortCircuit).Then(r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      chain,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
