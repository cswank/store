package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/GeertJohan/go.rice"
	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/handlers"
	"github.com/cswank/store/internal/shopify"
	"github.com/cswank/store/internal/site"
	"github.com/cswank/store/internal/storage"
	"github.com/cswank/store/internal/storage/mock"
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

	serve         = kingpin.Command("serve", "Start the server.")
	fakeRecaptcha = serve.Flag("fake-recaptcha", "start a fake shopify").Bool()

	website  = kingpin.Command("site", "manage site")
	generate = website.Command("generate", "generate a site")
	fake     = generate.Flag("fake-shopify", "start a fake shopify").Short('f').Bool()

	ts *httptest.Server
)

func doInit() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	box := rice.MustFindBox("templates")
	site.Init(cfg)
	templates.Init(cfg, box)
	handlers.Init(cfg)
	shopify.Init(cfg)
}

func main() {
	doInit()
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("Craig Swank")
	switch kingpin.Parse() {
	case "serve":
		templates.InitProducts()
		doServe()
	case "site generate":
		if *fake {
			cfg = shopify.FakeShopify()
			site.Init(cfg)
			shopify.Init(cfg)
		}
		err := site.Generate()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func doServe() {
	cfg = initServe()
	storage.Init(cfg)
	handlers.Init(cfg)

	r := mux.NewRouter().StrictSlash(true)

	r.Handle("/login", getMiddleware(handlers.Anyone, handlers.Login)).Methods("GET")
	r.Handle("/login", getMiddleware(handlers.Human, handlers.DoLogin)).Methods("POST")
	r.Handle("/logout", getMiddleware(handlers.Anyone, handlers.Logout)).Methods("GET")
	r.Handle("/logout", getMiddleware(handlers.Anyone, handlers.DoLogout)).Methods("POST")

	r.Handle("/wholesale", getMiddleware(handlers.Anyone, handlers.Wholesale)).Methods("GET")
	r.Handle("/wholesale", getMiddleware(handlers.Wholesaler, handlers.Purchase)).Methods("POST")
	r.Handle("/wholesale/register", getMiddleware(handlers.Human, handlers.WholesalerRegistration)).Methods("POST")
	r.Handle("/contact", getMiddleware(handlers.Human, handlers.Contact)).Methods("POST")
	r.Handle("/cart/lineitem/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.LineItem)).Methods("GET")
	r.PathPrefix("/").Handler(handlers.HandleErr(handlers.Static()))

	chain := alice.New(handlers.Log()).Then(r)
	addr := fmt.Sprintf("%s:%d", cfg.Iface, cfg.Port)

	var serve func() error

	server := &http.Server{
		Addr:         addr,
		Handler:      chain,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 120 * time.Second,
		//IdleTimeout:  120 * time.Second,
	}

	serve = server.ListenAndServe

	if cfg.UseTLS {
		serve = getTLS(server)
	}

	log.Printf("listening on %s (tls: %v)\n", addr, cfg.UseTLS)
	log.Println(serve())
}

func getTLS(srv *http.Server) func() error {
	if cfg.TLSCerts == "" {
		log.Fatal("you must set STORE_CERTS path when using tls")
	}
	if cfg.UseLetsEncrypt {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.Domains...),
			Cache:      autocert.DirCache(cfg.TLSCerts),
		}
		srv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}
	} else {
		c := filepath.Join(cfg.TLSCerts, "cert.pem")
		k := filepath.Join(cfg.TLSCerts, "key.pem")
		cer, err := tls.LoadX509KeyPair(c, k)
		if err != nil {
			log.Fatal(err)
		}
		srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	}
	go http.ListenAndServe(":80", http.HandlerFunc(handlers.Redirect))

	return func() error {
		return srv.ListenAndServeTLS("", "")
	}
}

func getMiddleware(perm handlers.ACL, f handlers.HandlerFunc) http.Handler {
	return alice.New(handlers.Authentication, handlers.Perm(perm)).Then(handlers.HandleErr(f))
}

func initServe() config.Config {
	if *fakeRecaptcha {
		return mock.FakeRecaptcha(cfg)
	}
	return cfg
}
