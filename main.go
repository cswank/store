package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/GeertJohan/go.rice"
	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/handlers"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/utils"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"golang.org/x/crypto/acme/autocert"
)

var (
	cfg      store.Config
	serve    = kingpin.Command("serve", "Start the server.")
	items    = kingpin.Command("items", "save and delete items")
	itemAdd  = items.Command("add", "add an item")
	itemEdit = items.Command("edit", "edit items")
	users    = kingpin.Command("users", "save and delete users")
	userAdd  = users.Command("add", "add an item")
	userEdit = users.Command("edit", "edit users")
	box      *rice.Box

	certPath, keyPath, port string
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
		certPath = os.Getenv("STORE_CERTPATH")
		keyPath = os.Getenv("STORE_KEYPATH")
		port = os.Getenv("STORE_PORT")

		if certPath == "" || keyPath == "" {
			log.Fatal("you must set STORE_CERTPATH and STORE_KEYPATH")
		}

		if port == "" {
			log.Fatal("you must set STORE_PORT")
		}

		box = rice.MustFindBox("static")
		handlers.Init(box)
		Serve()
	case "users add":
		utils.AddUser(store.GetDB())
	case "users edit":
		utils.EditUser(store.GetDB())
	}
}

func getMiddleware(perm handlers.ACL, f http.HandlerFunc) http.Handler {
	return alice.New(handlers.Authentication, handlers.Perm(perm), handlers.Handle(f)).Then(http.HandlerFunc(handlers.Errors))
}

func Serve() {
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/", getMiddleware(handlers.Anyone, handlers.Home)).Methods("GET")
	r.Handle("/login", getMiddleware(handlers.Anyone, handlers.Login)).Methods("GET")
	r.Handle("/login", getMiddleware(handlers.Anyone, handlers.DoLogin)).Methods("POST")
	r.Handle("/logout", getMiddleware(handlers.Anyone, handlers.Logout)).Methods("POST")
	r.Handle("/shop", getMiddleware(handlers.Anyone, handlers.Shop)).Methods("GET")
	r.Handle("/shop/{category}", getMiddleware(handlers.Anyone, handlers.Category)).Methods("GET")
	r.Handle("/shop/{category}/{subcategory}", getMiddleware(handlers.Anyone, handlers.SubCategory)).Methods("GET")
	r.Handle("/shop/{category}/{subcategory}/{item}", getMiddleware(handlers.Anyone, handlers.Item)).Methods("GET")
	r.Handle("/admin/items", getMiddleware(handlers.Admin, handlers.AdminPage)).Methods("GET")
	r.Handle("/admin/items", getMiddleware(handlers.Admin, handlers.AddItems)).Methods("POST")

	r.Handle("/favicon.ico", getMiddleware(handlers.Anyone, handlers.Favicon))

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(rice.MustFindBox("static").HTTPBox())))
	r.PathPrefix("/items/").Handler(http.StripPrefix("/items/", http.FileServer(rice.MustFindBox("internal/store/fixtures/items").HTTPBox())))

	chain := alice.New(handlers.Log(cfg.LogOutput)).Then(r)
	addr := fmt.Sprintf(":%s", port)

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("bleh.zekjur.net"),
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      chain,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    &tls.Config{GetCertificate: m.GetCertificate},
	}
	log.Printf("listening on %s\n", addr)
	log.Println(srv.ListenAndServeTLS("", ""))
}
