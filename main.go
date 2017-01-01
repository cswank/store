package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/GeertJohan/go.rice"
	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/handlers"
	"github.com/cswank/store/internal/site"
	"github.com/cswank/store/internal/storage"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	version = "0.0.0"
)

var (
	cfg config.Config

	serve    = kingpin.Command("serve", "Start the server.")
	fake     = serve.Flag("fake-shopify", "start a fake shopify").Short('f').Bool()
	website  = kingpin.Command("site", "manage site")
	generate = website.Command("generate", "generate a site")

	ts *httptest.Server
)

func doInit() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	box := rice.MustFindBox("templates")
	site.Init(cfg)
	templates.Init(box)
	handlers.Init(cfg)
}

func main() {
	doInit()
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("Craig Swank")
	switch kingpin.Parse() {
	case "serve":
		doServe()
	case "site generate":
		err := site.Generate()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func doServe() {
	storage.Init(cfg)
	r := mux.NewRouter().StrictSlash(true)

	r.Handle("/cart/lineitem/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.LineItem)).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

	chain := alice.New(handlers.Log(cfg.LogOutput)).Then(r)
	addr := fmt.Sprintf("%s:%d", cfg.Iface, cfg.Port)

	var serve func() error

	srv := &http.Server{
		Addr:         addr,
		Handler:      chain,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		//IdleTimeout:  120 * time.Second,  TODO uncomment when 1.8 is out
	}

	serve = srv.ListenAndServe

	if cfg.UseTLS {
		if cfg.TLSCerts == "" {
			log.Fatal("you must set STORE_CERTS path when using tls")
		}
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.Domains...),
			Cache:      autocert.DirCache(cfg.TLSCerts),
		}
		srv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		serve = func() error {
			return srv.ListenAndServeTLS("", "")
		}
		go http.ListenAndServe(":80", http.HandlerFunc(handlers.Redirect))
	}

	log.Printf("listening on %s (tls: %v)\n", addr, cfg.UseTLS)
	log.Println(serve())
}

func getMiddleware(perm handlers.ACL, f handlers.HandlerFunc) http.Handler {
	return alice.New(handlers.Authentication, handlers.Perm(perm)).Then(handlers.HandleErr(f))
}
