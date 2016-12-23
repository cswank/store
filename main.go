package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/GeertJohan/go.rice"
	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/handlers"
	"github.com/cswank/store/internal/shopify"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/utils"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"golang.org/x/crypto/acme/autocert"
)

var (
	cfg      store.Config
	serve    = kingpin.Command("serve", "Start the server.")
	fake     = serve.Flag("fake-shopify", "start a fake shopify").Short('f').Bool()
	items    = kingpin.Command("items", "save and delete items")
	itemAdd  = items.Command("add", "add an item")
	itemEdit = items.Command("edit", "edit items")

	categories = kingpin.Command("categories", "save, edit and delete categories")

	users             = kingpin.Command("users", "save and delete users")
	userAdd           = users.Command("add", "add an item")
	userEdit          = users.Command("edit", "edit users")
	products          = kingpin.Command("products", "save and delete products")
	deleteAllProducts = products.Command("delete-all", "edit users")
	box               *rice.Box

	port    string
	domains []string
	ts      *httptest.Server
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
		doServe()
	case "categories":
		utils.EditCategory()
	case "users add":
		utils.AddUser()
	case "users edit":
		utils.EditUser()
	}
}

func initServe() {
	if *fake {
		id := 1
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				var m map[string]shopify.Product
				json.NewDecoder(r.Body).Decode(&m)
				p := m["product"]
				p.ID = id
				p.Variants = []shopify.Variant{
					{ID: id},
				}
				m["product"] = p
				id++
				json.NewEncoder(w).Encode(m)
			}
		}))
		fmt.Println("ts", ts.URL)
		os.Setenv("SHOPIFY_DOMAIN", ts.URL)
		os.Setenv("SHOPIFY_API", ts.URL)
	}
	domain := os.Getenv("STORE_DOMAINS")
	if domain == "" {
		log.Println("tls not enabled because STORE_DOMAIN is not set")
	} else {
		domains = strings.Split(domain, ",")
	}

	port = os.Getenv("STORE_PORT")
	if port == "" {
		log.Fatal("you must set STORE_PORT")
	}

	box = rice.MustFindBox("static")
	shopify.Init()
	handlers.Init(box)
}

func getMiddleware(perm handlers.ACL, f handlers.HandlerFunc) http.Handler {
	return alice.New(handlers.Authentication, handlers.Perm(perm)).Then(handlers.HandleErr(f))
}

func doServe() {
	initServe()
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/", getMiddleware(handlers.Anyone, handlers.Home)).Methods("GET")
	r.Handle("/login", getMiddleware(handlers.Anyone, handlers.Login)).Methods("GET")
	r.Handle("/login", getMiddleware(handlers.Anyone, handlers.DoLogin)).Methods("POST")
	r.Handle("/logout", getMiddleware(handlers.Anyone, handlers.Logout)).Methods("GET")
	r.Handle("/logout", getMiddleware(handlers.Anyone, handlers.DoLogout)).Methods("POST")
	r.Handle("/contact", getMiddleware(handlers.Anyone, handlers.Contact)).Methods("GET")
	r.Handle("/contact", getMiddleware(handlers.Anyone, handlers.DoContact)).Methods("POST")
	r.Handle("/wholesale", getMiddleware(handlers.Anyone, handlers.Wholesale)).Methods("GET")
	r.Handle("/images/{type}/{title}/{size}", getMiddleware(handlers.Anyone, handlers.Image)).Methods("GET")
	r.Handle("/images/site/{title}", getMiddleware(handlers.Anyone, handlers.SiteImage)).Methods("GET")

	r.Handle("/cart", getMiddleware(handlers.Anyone, handlers.Cart)).Methods("GET")
	r.Handle("/cart/lineitem/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.LineItem)).Methods("GET")
	r.Handle("/shop", getMiddleware(handlers.Anyone, handlers.Shop)).Methods("GET")
	r.Handle("/shop/{category}", getMiddleware(handlers.Anyone, handlers.Category)).Methods("GET")
	r.Handle("/shop/{category}/{subcategory}", getMiddleware(handlers.Anyone, handlers.SubCategory)).Methods("GET")
	r.Handle("/shop/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.Product)).Methods("GET")

	r.Handle("/api/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.GetProduct)).Methods("GET")

	r.Handle("/admin", getMiddleware(handlers.Admin, handlers.AdminPage)).Methods("GET")
	r.Handle("/admin/db/backup", getMiddleware(handlers.Admin, handlers.BackupDB)).Methods("GET")
	r.Handle("/admin/confirm", getMiddleware(handlers.Admin, handlers.Confirm)).Methods("GET")
	r.Handle("/admin/site/images/home", getMiddleware(handlers.Admin, handlers.AddHomeImage)).Methods("POST")
	r.Handle("/admin/categories", getMiddleware(handlers.Admin, handlers.AddCategory)).Methods("POST")
	r.Handle("/admin/categories/{category}", getMiddleware(handlers.Admin, handlers.AdminCategoryPage)).Methods("GET")
	r.Handle("/admin/categories/{category}", getMiddleware(handlers.Admin, handlers.RenameCategory)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories", getMiddleware(handlers.Admin, handlers.AddCategory)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}", getMiddleware(handlers.Admin, handlers.AdminAddProductPage)).Methods("GET")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}", getMiddleware(handlers.Admin, handlers.RenameSubcategory)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products", getMiddleware(handlers.Admin, handlers.AddProduct)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products/{title}", getMiddleware(handlers.Admin, handlers.AdminProductPage)).Methods("GET")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products/{title}", getMiddleware(handlers.Admin, handlers.UpdateProduct)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products/{title}", getMiddleware(handlers.Admin, handlers.DeleteProduct)).Methods("DELETE")

	r.Handle("/favicon.ico", getMiddleware(handlers.Anyone, handlers.Favicon))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box.HTTPBox())))

	chain := alice.New(handlers.Log(cfg.LogOutput)).Then(r)
	iface := os.Getenv("STORE_IFACE")
	addr := fmt.Sprintf("%s:%s", iface, port)

	var serve func() error

	srv := &http.Server{
		Addr:         addr,
		Handler:      chain,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		//IdleTimeout:  120 * time.Second,  TODO uncomment when 1.8 is out
	}

	serve = srv.ListenAndServe

	var withTLS bool

	if os.Getenv("STORE_TLS") == "true" {
		certs := os.Getenv("STORE_CERTS")
		if certs == "" {
			log.Fatal("you must set STORE_CERTS path when using tls")
		}
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
			Cache:      autocert.DirCache(certs),
		}
		srv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		serve = func() error {
			return srv.ListenAndServeTLS("", "")
		}
		withTLS = true
	}

	if withTLS {
		go http.ListenAndServe(":80", http.HandlerFunc(handlers.Redirect))
	}

	log.Printf("listening on %s (tls: %v)\n", addr, withTLS)
	log.Println(serve())
}
