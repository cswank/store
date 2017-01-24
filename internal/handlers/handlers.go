package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	rice "github.com/GeertJohan/go.rice"
	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
)

var (
	errInvalidLogin = errors.New("invalid login")
	lock            sync.Mutex
	shoppingLinks   []link
	cfg             config.Config
	box             *rice.Box
	ico             []byte
)

func Init(c config.Config, b *rice.Box) {
	cfg = c
	box = b
	shopify = shopifyAPI{
		APIKey: cfg.ShopifyJSKey,
		Domain: cfg.ShopifyDomain,
	}

	if shopify.APIKey == "" || shopify.Domain == "" {
		log.Fatal("you must set SHOPIFY_DOMAIN and SHOPIFY_JS_KEY")
	}

	if cfg.RecaptchaSiteKey != "" && cfg.RecaptchaSecretKey != "" && cfg.RecaptchaURL != "" {
		captcha = true
	}

	storeEmail = cfg.Email
	storeEmailPassword = cfg.EmailPassword
	if storeEmail == "" || storeEmailPassword == "" {
		log.Fatal("you must set STORE_EMAIL and STORE_EMAIL_PASSWORD")
	}

	makeNavbarLinks()
	etags = make(map[string]string)
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
		// {Name: "Wholesale", Link: "/wholesale"},
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

	cats, err := store.GetCategories()
	if err != nil {
		lg.Println("error getting cats")
		return nil
	}

	var l []link

	for _, cat := range cats {
		l = append(l, link{Category: true, Name: cat})
		l = append(l, getSubcatLinks(cat)...)
	}

	shoppingLinks = l
	return shoppingLinks
}

func getSubcatLinks(cat string) []link {
	subcats, err := store.GetSubCategories(cat)
	if err != nil {
		lg.Println("error getting subcats")
		return nil
	}

	l := make([]link, len(subcats))

	for i, subcat := range subcats {
		l[i] = link{
			Link: fmt.Sprintf("/shop/%s/%s", cat, subcat),
			Name: subcat,
		}
	}

	return l
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
		Shopify: shopify,
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

func ServeBox(w http.ResponseWriter, req *http.Request) error {
	pth := req.URL.Path
	if strings.HasPrefix(pth, ".") || strings.HasPrefix(pth, "/") {
		return store.ErrNotFound
	}

	f, err := box.Open(pth)
	if err != nil {
		return err
	}

	if strings.Contains(pth, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}

	w.Header().Set("Cache-Control", "max-age=86400")
	io.Copy(w, f)
	f.Close()

	return nil
}
