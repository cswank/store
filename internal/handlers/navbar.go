package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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
		{Name: "Blog", Link: "/blog"},
		{Name: "About", Link: "/about"},
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
