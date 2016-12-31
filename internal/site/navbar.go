package site

import "fmt"

type link struct {
	Name     string
	Link     string
	Style    string
	Children []link
}

func getNavbarLinks(cats categories) []link {
	return []link{
		{Name: "Home", Link: "/"},
		{Name: "Shop", Link: "/", Children: getShoppingLinks(cats)},
		{Name: "Contact", Link: "/contact"},
		{Name: "Cart", Link: "/cart"},
	}
}

func getShoppingLinks(cats categories) []link {
	var l []link

	for cat, subcats := range cats {
		l = append(l, getSubcatLinks(cat, subcats)...)
	}

	return l
}

func getSubcatLinks(cat string, subcats map[string][]string) []link {

	l := make([]link, len(subcats))

	var i int
	for subcat := range subcats {
		l[i] = link{
			Link: fmt.Sprintf("/products/%s/%s", cat, subcat),
			Name: subcat,
		}
		i++
	}

	return l
}
