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
	"github.com/cswank/store/internal/utils"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
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
	r.Handle("/store/items", getMiddleware(handlers.Anyone, handlers.Items)).Methods("GET")
	r.Handle("/store/items/{category}/{subcategory}", getMiddleware(handlers.Anyone, handlers.SubCategory)).Methods("GET")
	r.Handle("/store/items/{category}/{subcategory}/{item}", getMiddleware(handlers.Anyone, handlers.Item)).Methods("GET")
	r.Handle("/admin/items", getMiddleware(handlers.Admin, handlers.AdminPage)).Methods("GET")
	r.Handle("/admin/items", getMiddleware(handlers.Admin, handlers.AddItems)).Methods("POST")

	r.Handle("/favicon.ico", getMiddleware(handlers.Anyone, handlers.Favicon))

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(rice.MustFindBox("static").HTTPBox())))
	r.PathPrefix("/items/").Handler(http.StripPrefix("/items/", http.FileServer(rice.MustFindBox("internal/store/fixtures/items").HTTPBox())))

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
