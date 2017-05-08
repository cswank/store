package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/securecookie"
)

var (
	errInvalidLogin = errors.New("invalid login")
	lock            sync.Mutex
	shoppingLinks   []link
	cfg             config.Config
	ico             []byte

	//holds external html that is injected into the app
	html map[string]template.HTML
)

func Init(c config.Config) {
	cfg = c
	shopifyKey = shopifyAPI{
		APIKey: cfg.ShopifyJSKey,
		Domain: cfg.ShopifyDomain,
	}

	htmlLookup := map[string]string{
		"about": cfg.About,
		"head":  cfg.Head,
		"home":  cfg.Home,
	}

	html = make(map[string]template.HTML)
	for name, pth := range htmlLookup {
		d, err := ioutil.ReadFile(pth)
		if err != nil {
			log.Fatal("could not read ", name, pth)
		}
		html[name] = template.HTML(string(d))
	}

	if shopifyKey.APIKey == "" || shopifyKey.Domain == "" {
		log.Fatal("you must set SHOPIFY_DOMAIN and SHOPIFY_JS_KEY")
	}

	if cfg.RecaptchaSiteKey != "" && cfg.RecaptchaSecretKey != "" && cfg.RecaptchaURL != "" {
		captcha = true
	}

	makeNavbarLinks()
	etags = make(map[string]string)

	domains := cfg.Domains
	if len(domains) == 0 {
		log.Fatal("you must set STORE_DOMAINS")
	}

	domain = domains[0]
	authCookieName = fmt.Sprintf("%s-user", domain)
	hashKey = []byte(cfg.HashKey)
	blockKey = []byte(cfg.BlockKey)
	if string(hashKey) == "" || string(blockKey) == "" {
		log.Fatal("you must set STORE_HASH_KEY and STORE_BLOCK_KEY")
	}
	sc = securecookie.New(hashKey, blockKey)
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type link struct {
	Category bool
	Name     string
	Link     string
	Style    string
	Children []link
}

func getNavbarLinks(req *http.Request) []link {
	l := []link{
		{Name: "Home", Link: "/"},
		{Name: "Shop", Link: "/", Children: getShoppingLinks()},
		{Name: "Blog", Link: "/blog"},
		{Name: "About Me", Link: "/about"},
		//{Name: "Wholesale", Link: "/wholesale"},
		{Name: "Contact", Link: "/contact"},
		{Name: "Cart", Link: "/cart"},
	}

	if Admin(req) {
		l = append(l, link{Name: "Admin", Link: "/admin"})
	}

	if Read(req) {
		l = append(l, link{Name: "Logout", Link: "/logout", Style: "float:right"})
	}

	return l
}

func makeNavbarLinks() {
	lock.Lock()
	shoppingLinks = nil
	lock.Unlock()
	getShoppingLinks()
}

func getShoppingLinks() []link {
	lock.Lock()
	defer lock.Unlock()

	if shoppingLinks != nil {
		return shoppingLinks
	}

	f, err := os.Open(cfg.ShoppingMenu)
	if err != nil {
		log.Fatal("could not open shopping menu", cfg.ShoppingMenu, err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&shoppingLinks); err != nil {
		log.Fatal("could not decode shopping menu", err)
	}

	return shoppingLinks
}

func Static() HandlerFunc {
	srv := http.FileServer(http.Dir("."))
	return func(w http.ResponseWriter, req *http.Request) error {
		// pusher, ok := w.(http.Pusher)
		// if ok {
		// 	for _, resource := range pushes[req.URL.Path] {
		// 		if err := pusher.Push(resource, nil); err != nil {
		// 			return err
		// 		}
		// 	}
		// }
		srv.ServeHTTP(w, req)
		return nil
	}
}

func HandleErr(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		if err == errInvalidLogin {
			handleInvalidLogin(w)
		} else if err == store.ErrNotFound {
			NotFound(w, r)
		} else {
			lg.Println("internal server err", r.URL.RawPath, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func NotFound(w http.ResponseWriter, req *http.Request) {
	p := page{
		Links:   getNavbarLinks(req),
		Admin:   Admin(req),
		Shopify: shopifyKey,
		Name:    name,
	}
	templates.Get("notfound.html").ExecuteTemplate(w, "base", p)
}

func handleInvalidLogin(w http.ResponseWriter) {
	w.Header().Set("Location", "/login?error=email or password is incorrect, please try again")
	w.WriteHeader(http.StatusFound)
}

// func Favicon(w http.ResponseWriter, req *http.Request) error {
// 	w.Header().Set("Cache-Control", "max-age=86400")
// 	w.Write(ico)
// 	return nil
// }
