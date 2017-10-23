package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cswank/store/internal/store"
)

type link struct {
	Category bool
	Name     string
	Link     string
	HasLink  bool
	Style    string
	Children []link
}

func getNavbarLinks(req *http.Request) []link {
	l := []link{
		{Name: "Home", Link: "/"},
		{Name: "Shop", Link: "/", Children: getShoppingLinks()},
		//{Name: "Blog", Link: "/blog"},
		//{Name: "About", Link: "/about"},
		{Name: "Wholesale", Link: "/wholesale"},
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
	return getDBShoppingLinks()
}

func getMenuShoppingLinks() []link {
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

func getDBShoppingLinks() []link {
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
		scl := getSubcatLinks(cat)
		l = append(l, link{Category: true, Name: cat})
		if len(scl) > 0 {
			l = append(l, scl...)
		}
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
		if subcat == "NOSUBCATEGORIES" {
			return []link{}
		}

		l[i] = link{
			Link: fmt.Sprintf("/shop/%s/%s", cat, subcat),
			Name: subcat,
		}
	}

	return l
}
