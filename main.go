package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/GeertJohan/go.rice"
	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/email"
	"github.com/cswank/store/internal/handlers"
	"github.com/cswank/store/internal/shopify"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/cswank/store/internal/utils"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"golang.org/x/crypto/acme/autocert"
)

var (
	cfg      config.Config
	serve    = kingpin.Command("serve", "Start the server.")
	fake     = serve.Flag("fake-shopify", "start a fake shopify").Short('f').Bool()
	items    = kingpin.Command("items", "save and delete items")
	itemAdd  = items.Command("add", "add an item")
	itemEdit = items.Command("edit", "edit items")

	users    = kingpin.Command("users", "save and delete users")
	userAdd  = users.Command("add", "add an item")
	userEdit = users.Command("edit", "edit users")

	categories = kingpin.Command("categories", "save, edit and delete categories")
	edit       = categories.Command("edit", "edit categories and subcategories")

	blogs = kingpin.Command("blogs", "save, edit and delete blogs")
	_     = blogs.Command("edit", "edit a blog")

	box       *rice.Box
	staticBox *rice.Box

	ts *httptest.Server
)

const (
	version = "0.0.0"
)

func init() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("could not parse config", err)
	}
	store.Init(cfg)
	email.Init(cfg)
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
	case "categories edit":
		utils.EditCategory()
	case "blogs edit":
		utils.EditBlog()
	}
}

func initServe() {
	if *fake {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				if strings.Contains(r.URL.Path, "products") {
					var m map[string]shopify.Product
					json.NewDecoder(r.Body).Decode(&m)
					p := m["product"]
					p.ID = rand.Int()
					p.Variants = []shopify.Variant{
						{ID: p.ID},
					}
					m["product"] = p
					json.NewEncoder(w).Encode(m)
				} else if strings.Contains(r.URL.Path, "discounts") {
					var m map[string]shopify.DiscountCode
					json.NewDecoder(r.Body).Decode(&m)
					dc := m["discount"]
					dc.ID = int(rand.Int63())
					m["discount"] = dc
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(m)
				}
			}
		}))
		//cfg.Domains = []string{ts.URL}
		cfg.ShopifyAPI = ts.URL
	}

	box = rice.MustFindBox("templates")
	shopify.Init(cfg)
	handlers.Init(cfg)
	templates.Init(box)
}

func getMiddleware(perm handlers.ACL, f handlers.HandlerFunc) http.Handler {
	return alice.New(handlers.Authentication, handlers.Perm(perm)).Then(handlers.HandleErr(f))
}

func getImageMiddleware(perm handlers.ACL, f handlers.HandlerFunc) http.Handler {
	return alice.New(handlers.ETag, handlers.Authentication, handlers.Perm(perm)).Then(handlers.HandleErr(f))
}

func doServe() {
	initServe()

	restartChan := make(chan bool)

	r := mux.NewRouter().StrictSlash(true)

	if cfg.WebhookID != "" && cfg.WebhookScript != "" && cfg.WebhookIPWhitelist != "" {
		r.Handle("/webhooks/{id}", getMiddleware(handlers.IPWhitelist, handlers.GetWebhooks(restartChan))).Methods("POST")
	}

	r.Handle("/", getMiddleware(handlers.Anyone, handlers.Home)).Methods("GET")
	r.Handle("/login", getMiddleware(handlers.Anyone, handlers.Login)).Methods("GET")
	r.Handle("/login", getMiddleware(handlers.Human, handlers.DoLogin)).Methods("POST")
	r.Handle("/login/reset", getMiddleware(handlers.Anyone, handlers.ResetPage)).Methods("GET")
	r.Handle("/login/reset", getMiddleware(handlers.Human, handlers.SendReset)).Methods("POST")
	r.Handle("/login/do-reset", getMiddleware(handlers.Anyone, handlers.ResetPassword)).Methods("GET")
	r.Handle("/login/do-reset", getMiddleware(handlers.Anyone, handlers.DoResetPassword)).Methods("POST")
	r.Handle("/logout", getMiddleware(handlers.Anyone, handlers.Logout)).Methods("GET")
	r.Handle("/logout", getMiddleware(handlers.Anyone, handlers.DoLogout)).Methods("POST")

	r.Handle("/contact", getMiddleware(handlers.Anyone, handlers.Contact)).Methods("GET")
	r.Handle("/contact", getMiddleware(handlers.Human, handlers.DoContact)).Methods("POST")

	r.Handle("/wholesale", getMiddleware(handlers.Anyone, handlers.Wholesale)).Methods("GET")
	r.Handle("/wholesale/invoice", getMiddleware(handlers.Wholesaler, handlers.Invoice)).Methods("GET", "POST")
	r.Handle("/wholesale/application", getMiddleware(handlers.Anyone, handlers.WholesaleApplication)).Methods("GET")
	r.Handle("/wholesale/application", getMiddleware(handlers.Anyone, handlers.WholesaleApply)).Methods("POST")
	r.Handle("/wholesale/application/{token}", getMiddleware(handlers.Anyone, handlers.WholesaleVerify)).Methods("GET")
	r.Handle("/wholesale/thanks", getMiddleware(handlers.Anyone, handlers.WholesaleThanks)).Methods("GET")

	r.Handle("/cart", getMiddleware(handlers.Anyone, handlers.Cart)).Methods("GET")
	r.Handle("/cart/lineitem/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.LineItem)).Methods("GET")

	r.Handle("/blog", getMiddleware(handlers.Anyone, handlers.Blog)).Methods("GET")
	r.Handle("/blog/{blog}", getMiddleware(handlers.Anyone, handlers.Blog)).Methods("GET")
	r.Handle("/images/blogs/{blog}", getMiddleware(handlers.Anyone, handlers.BlogImage)).Methods("GET")

	r.Handle("/about", getMiddleware(handlers.Anyone, handlers.About)).Methods("GET")

	r.Handle("/shop", getMiddleware(handlers.Anyone, handlers.Shop)).Methods("GET")
	r.Handle("/shop/{category}", getMiddleware(handlers.Anyone, handlers.Category)).Methods("GET")
	r.Handle("/shop/{category}/{subcategory}", getMiddleware(handlers.Anyone, handlers.SubCategory)).Methods("GET")
	r.Handle("/shop/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.Product)).Methods("GET")
	r.Handle("/shop/images/{type}/{title}/{size}", getImageMiddleware(handlers.Anyone, handlers.Image)).Methods("GET")

	r.Handle("/api/{category}/{subcategory}/{title}", getMiddleware(handlers.Anyone, handlers.GetProduct)).Methods("GET")

	r.Handle("/admin", getMiddleware(handlers.Admin, handlers.AdminPage)).Methods("GET")

	r.Handle("/admin/blogs", getMiddleware(handlers.Admin, handlers.ManageBlogs)).Methods("GET")
	r.Handle("/admin/blogs", getMiddleware(handlers.Admin, handlers.CreateBlog)).Methods("POST")
	r.Handle("/admin/blogs/{blog}", getMiddleware(handlers.Admin, handlers.BlogForm)).Methods("GET")
	r.Handle("/admin/blogs/{blog}", getMiddleware(handlers.Admin, handlers.UpdateBlog)).Methods("POST")
	r.Handle("/admin/blogs/{blog}", getMiddleware(handlers.Admin, handlers.DeleteBlog)).Methods("DELETE")
	r.Handle("/admin/wholesalers", getMiddleware(handlers.Admin, handlers.AdminWholesalers)).Methods("GET")
	r.Handle("/admin/wholesalers/{wholesaler}", getMiddleware(handlers.Admin, handlers.AdminWholesaler)).Methods("GET")
	r.Handle("/admin/wholesalers/{wholesaler}", getMiddleware(handlers.Admin, handlers.AdminWholesalerUpdate)).Methods("POST")
	r.Handle("/admin/wholesalers/{wholesaler}", getMiddleware(handlers.Admin, handlers.AdminWholesalerDelete)).Methods("DELETE")
	r.Handle("/admin/wholesalers/{wholesaler}/confirmation", getMiddleware(handlers.Admin, handlers.AdminWholesalerConfirm)).Methods("POST")
	r.Handle("/admin/db/backup", getMiddleware(handlers.Admin, handlers.BackupDB)).Methods("GET")
	r.Handle("/admin/confirm", getMiddleware(handlers.Admin, handlers.Confirm)).Methods("GET")
	r.Handle("/admin/categories", getMiddleware(handlers.Admin, handlers.AddCategory)).Methods("POST")
	r.Handle("/admin/categories/{category}", getMiddleware(handlers.Admin, handlers.AdminCategoryPage)).Methods("GET")
	r.Handle("/admin/categories/{category}", getMiddleware(handlers.Admin, handlers.RenameCategory)).Methods("POST")
	r.Handle("/admin/categories/{category}", getMiddleware(handlers.Admin, handlers.DeleteCategory)).Methods("DELETE")
	r.Handle("/admin/categories/{category}/price", getMiddleware(handlers.Admin, handlers.UpdatePrice)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories", getMiddleware(handlers.Admin, handlers.AddCategory)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}", getMiddleware(handlers.Admin, handlers.AdminAddProductPage)).Methods("GET")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}", getMiddleware(handlers.Admin, handlers.RenameSubcategory)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products", getMiddleware(handlers.Admin, handlers.AddProduct)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products/{title}", getMiddleware(handlers.Admin, handlers.AdminProductPage)).Methods("GET")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products/{title}", getMiddleware(handlers.Admin, handlers.UpdateProduct)).Methods("POST")
	r.Handle("/admin/categories/{category}/subcategories/{subcategory}/products/{title}", getMiddleware(handlers.Admin, handlers.DeleteProduct)).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(handlers.NotFound)

	r.PathPrefix("/images").Handler(handlers.HandleErr(handlers.Static()))
	r.PathPrefix("/robots.txt").Handler(handlers.HandleErr(handlers.Static()))
	r.PathPrefix("/favicon.ico").Handler(handlers.HandleErr(handlers.Static()))
	r.PathPrefix("/css").Handler(handlers.HandleErr(handlers.Static()))
	r.PathPrefix("/js").Handler(handlers.HandleErr(handlers.Static()))

	chain := alice.New(handlers.Log(cfg.LogOutput)).Then(r)
	iface := os.Getenv("STORE_IFACE")
	addr := fmt.Sprintf("%s:%d", iface, cfg.Port)

	var serve func() error

	srv := &http.Server{
		Addr:         addr,
		Handler:      chain,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		//IdleTimeout:  120 * time.Second,  TODO uncomment when 1.8 is out
	}

	serve = srv.ListenAndServe

	if cfg.TLS {
		serve = getTLS(srv)
	}

	log.Printf("listening on %s (tls: %v)\n", addr, cfg.TLS)
	go func() {
		log.Println(serve())
	}()

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	var restart bool
	select {
	case <-stopChan:
	case restart = <-restartChan:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	if !restart {
		log.Println("shutting down")
		return
	}

	log.Println("restarting")
	doServe()
}

func getTLS(srv *http.Server) func() error {
	if cfg.TLSCerts == "" {
		log.Fatal("you must set STORE_CERTS path when using tls")
	}

	if cfg.LetsEncrypt {
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
			log.Fatal("could not load tls certs", err)
		}
		srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	}
	go http.ListenAndServe(":80", http.HandlerFunc(handlers.Redirect))

	return func() error {
		return srv.ListenAndServeTLS("", "")
	}
}
