package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

type adminPage struct {
	page
	Title            string
	URI              string
	Placeholder      string
	From             string
	IsProduct        bool
	Product          *store.Product
	ProductID        string
	ProductTitle     string
	BackgroundImages []string
	Items            []string
	AdminLinks       []link
	Subcategories    []string
}

func AdminPage(w http.ResponseWriter, req *http.Request) error {
	categories, err := store.GetCategories()
	if err != nil {
		return err
	}

	p := adminPage{
		page: page{
			Admin:   Admin(getUser(req)),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
		Items:       categories,
		URI:         "/admin/categories",
		From:        "/admin",
		Placeholder: "new category",
		AdminLinks:  []link{{Name: "Categories", Link: "/admin"}},
	}
	return templates["admin.html"].template.ExecuteTemplate(w, "base", p)
}

func AddCategory(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	name := req.FormValue("Name")
	vars := mux.Vars(req)
	cat := vars["category"]
	if cat == "" {
		if err := store.AddCategory(name); err != nil {
			return err
		}
	} else {
		if err := store.AddSubCategory(cat, name); err != nil {
			return err
		}
		makeNavbarLinks()
	}

	from := req.URL.Query().Get("from")
	if from == "" {
		from = "/admin"
	}

	w.Header().Set("Location", from)
	w.WriteHeader(http.StatusFound)
	return nil
}

func AdminCategoryPage(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	subcats, err := store.GetSubCategories(vars["category"])
	if err != nil {
		return err
	}

	from := fmt.Sprintf("/admin/categories/%s", vars["category"])
	p := adminPage{
		page: page{
			Admin:   Admin(getUser(req)),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
		Items:       subcats,
		From:        from,
		URI:         fmt.Sprintf("/admin/categories/%s/subcategories", vars["category"]),
		Placeholder: "new sub-category",
		AdminLinks: []link{
			{Name: "Categories", Link: "/admin"},
			{Name: vars["category"], Link: from},
		},
	}
	return templates["admin.html"].template.ExecuteTemplate(w, "base", p)
}

func AddProduct(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	cat := vars["category"]
	subcat := vars["subcategory"]

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	ff, _, err := req.FormFile("Image")
	if err != nil {
		return err
	}
	defer ff.Close()
	name := req.FormValue("Name")
	description := strings.Replace(req.FormValue("Description"), "\n", "", -1)

	p := store.NewProduct(name, cat, subcat, description)
	err = p.Add(ff)
	if err != nil {
		return err
	}

	l := fmt.Sprintf("/admin/categories/%s/subcategories/%s", cat, subcat)
	w.Header().Set("Location", l)
	w.WriteHeader(http.StatusFound)
	return nil
}

func AdminAddProductPage(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	products, err := store.GetProducts(vars["category"], vars["subcategory"])
	if err != nil {
		return err
	}

	from := fmt.Sprintf("/admin/categories/%s/subcategories/%s", vars["category"], vars["subcategory"])
	p := adminPage{
		page: page{
			Admin:   Admin(getUser(req)),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
		Items:       products,
		From:        from,
		URI:         fmt.Sprintf("/admin/categories/%s/subcategories/%s/products", vars["category"], vars["subcategory"]),
		Placeholder: "new product",
		AdminLinks: []link{
			{Name: "Categories", Link: "/admin"},
			{Name: vars["category"], Link: fmt.Sprintf("/admin/categories/%s", vars["category"])},
			{Name: vars["subcategory"], Link: from},
		},
		IsProduct: true,
	}

	return templates["admin.html"].template.ExecuteTemplate(w, "base", p)
}

func AdminProductPage(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	p := store.NewProduct(vars["title"], vars["category"], vars["subcategory"], "")
	err := p.Fetch()
	if err != nil {
		return err
	}

	subs, err := store.GetSubCategories(p.Cat)
	if err != nil {
		return err
	}

	from := url.QueryEscape(fmt.Sprintf("/admin/categories/%s/subcategories/%s/products/%s", p.Cat, p.Subcat, p.Title))
	page := adminPage{
		page: page{
			Admin:   Admin(getUser(req)),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
		From:        from,
		URI:         from,
		Placeholder: "new product",
		AdminLinks: []link{
			{Name: "Categories", Link: "/admin"},
			{Name: p.Cat, Link: fmt.Sprintf("/admin/categories/%s", p.Cat)},
			{Name: p.Subcat, Link: fmt.Sprintf("/admin/categories/%s/subcategories/%s", p.Cat, p.Subcat)},
			{Name: p.Title, Link: from},
		},
		Product:       p,
		Subcategories: subs,
	}

	return templates["admin-product.html"].template.ExecuteTemplate(w, "base", page)
}

func DeleteProduct(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	p := store.NewProduct(vars["title"], vars["category"], vars["subcategory"], "")
	if err := p.Fetch(); err != nil {
		return err
	}

	if err := p.Delete(); err != nil {
		return err
	}

	return nil
}

func UpdateProduct(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	p := store.NewProduct(vars["title"], vars["category"], vars["subcategory"], "")
	err := p.Fetch()
	if err != nil {
		return err
	}

	if err := req.ParseForm(); err != nil {
		return err
	}

	title := req.FormValue("Title")
	sub := req.FormValue("Subcat")
	p2 := store.NewProduct(title, p.Cat, sub, p.Description)

	if err := p.Update(p2); err != nil {
		return err
	}

	makeNavbarLinks()
	l := fmt.Sprintf("/admin/categories/%s/subcategories/%s/products/%s", p.Cat, p2.Subcat, p2.Title)
	w.Header().Set("Location", l)
	w.WriteHeader(http.StatusFound)

	return nil
}

type confirmPage struct {
	page
	Name     string
	Resource string
}

func Confirm(w http.ResponseWriter, req *http.Request) error {
	args := req.URL.Query()
	name := args.Get("name")
	resource := args.Get("resource")
	p := confirmPage{
		Name:     name,
		Resource: resource,
	}
	return templates["confirm.html"].template.ExecuteTemplate(w, "base", p)
}
