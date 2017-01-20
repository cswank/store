package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/cswank/store/internal/store"
)

var (
	errInvalidLogin = errors.New("invalid login")
	lock            sync.Mutex
	shoppingLinks   []link
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type link struct {
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

func HandleErr(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		if err == errInvalidLogin {
			handleInvalidLogin(w)
		} else if err == store.ErrNotFound {
			handleNotFound(w)
		} else {
			lg.Println("internal server err", r.URL.RawPath, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func handleNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

func handleInvalidLogin(w http.ResponseWriter) {
	w.Header().Set("Location", "/login?error=email or password is incorrect, please try again")
	w.WriteHeader(http.StatusFound)
}
