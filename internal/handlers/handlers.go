package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/storage"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var (
	cfg                    config.Config
	sc                     *securecookie.SecureCookie
	domain, authCookieName string
	errInvalidLogin        = errors.New("invalid login")
	pushes                 map[string][]string
)

func Init(c config.Config) {
	cfg = c
	fmt.Println("pushes", cfg.Domains)

	if err := json.Unmarshal([]byte(cfg.Pushes), &pushes); err != nil {
		log.Println("no pushes defined in config ", err)
	}

	domain = cfg.Domains[0]
	authCookieName = fmt.Sprintf("%s-user", domain)
	sc = securecookie.New([]byte(cfg.CookieHashKey), []byte(cfg.CookieBlockKey))
}

func SetConfig(c config.Config) {
	cfg = c
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

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

func LineItem(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	p := templates.NewProduct(vars["title"], vars["category"], vars["subcategory"])
	vals := req.URL.Query()
	p.ID = vals.Get("id")

	qs := vals.Get("quantity")
	if qs == "" {
		return fmt.Errorf("you must supply a quantity")
	}

	q, err := strconv.ParseInt(qs, 10, 64)
	if err != nil {
		return err
	}

	p.Quantity = int(q)
	price, err := strconv.ParseFloat(cfg.DefaultPrice, 10)
	if err != nil {
		return err
	}

	t := float64(q) * price
	p.Total = fmt.Sprintf("%.02f", t)
	return templates.Get("lineitem.html").ExecuteTemplate(w, "lineitem.html", p)
}

func Redirect(w http.ResponseWriter, req *http.Request) {
	http.Redirect(
		w,
		req,
		fmt.Sprintf("https://%s%s", req.Host, req.URL.String()),
		http.StatusMovedPermanently,
	)
}

func HandleErr(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		if err == errInvalidLogin {
			handleInvalidLogin(w)
		} else if err == storage.ErrNotFound {
			handleNotFound(w)
		} else {
			log.Println("internal server err", r.URL.RawPath, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func handleNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

func handleInvalidLogin(w http.ResponseWriter) {
	w.Header().Set("Location", "/login.html?error=invalid-login")
	w.WriteHeader(http.StatusFound)
}
