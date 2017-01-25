package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
)

type adminPage struct {
	page
	Title            string
	URI              string
	Resource         string
	ResourceName     string
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

func BackupDB(w http.ResponseWriter, req *http.Request) error {
	return store.GetBackup(w)
}

func AdminPage(w http.ResponseWriter, req *http.Request) error {
	categories, err := store.GetCategories()
	if err != nil {
		return err
	}

	p := adminPage{
		page: page{
			Admin:   Admin(req),
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
	return templates.Get("admin.html").ExecuteTemplate(w, "base", p)
}

func AddCategory(w http.ResponseWriter, req *http.Request) error {
	cat, _, _ := getVars(req)

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	name := req.FormValue("Name")
	if strings.Contains(name, "/") {
		return fmt.Errorf("illegal character (/)")
	}

	if cat == "" {
		if err := store.AddCategory(name); err != nil {
			return err
		}
	} else {
		if err := store.AddSubcategory(cat, name); err != nil {
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
	cat, _, _ := getVars(req)

	subcats, err := store.GetSubCategories(cat)
	if err != nil {
		return err
	}

	from := fmt.Sprintf("/admin/categories/%s", cat)
	p := adminPage{
		page: page{
			Admin:   Admin(req),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
		Items:        subcats,
		From:         from,
		URI:          fmt.Sprintf("/admin/categories/%s/subcategories", cat),
		Resource:     fmt.Sprintf("/admin/categories/%s", cat),
		ResourceName: cat,
		Placeholder:  "new sub-category",
		AdminLinks: []link{
			{Name: "Categories", Link: "/admin"},
			{Name: cat, Link: from},
		},
	}
	return templates.Get("admin.html").ExecuteTemplate(w, "base", p)
}

func RenameCategory(w http.ResponseWriter, req *http.Request) error {
	cat, _, _ := getVars(req)

	if err := req.ParseForm(); err != nil {
		return err
	}

	newName := req.FormValue("Name")
	err := store.RenameCategory(cat, newName)
	if err != nil {
		return err
	}

	makeNavbarLinks()
	l := fmt.Sprintf("/admin/categories/%s", newName)
	w.Header().Set("Location", l)
	w.WriteHeader(http.StatusFound)
	return nil
}

func RenameSubcategory(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, _ := getVars(req)

	if err := req.ParseForm(); err != nil {
		return err
	}

	newName := req.FormValue("Name")
	err := store.RenameSubcategory(cat, subcat, newName)
	if err != nil {
		return err
	}

	makeNavbarLinks()
	l := fmt.Sprintf("/admin/categories/%s/subcategories/%s", cat, newName)
	w.Header().Set("Location", l)
	w.WriteHeader(http.StatusFound)
	return nil
}

func AddProduct(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, _ := getVars(req)

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

	p := store.NewProduct(name, cat, subcat, store.ProductDescription(description))
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
	cat, subcat, _ := getVars(req)

	products, err := store.GetProducts(cat, subcat)
	if err != nil {
		return err
	}

	from := fmt.Sprintf("/admin/categories/%s/subcategories/%s", cat, subcat)
	p := adminPage{
		page: page{
			Admin:   Admin(req),
			Links:   getNavbarLinks(req),
			Shopify: shopify,
			Name:    name,
		},
		Items:        products,
		From:         from,
		URI:          fmt.Sprintf("/admin/categories/%s/subcategories/%s/products", cat, subcat),
		Resource:     fmt.Sprintf("/admin/categories/%s/subcategories/%s", cat, subcat),
		ResourceName: subcat,
		Placeholder:  "new product",
		AdminLinks: []link{
			{Name: "Categories", Link: "/admin"},
			{Name: cat, Link: fmt.Sprintf("/admin/categories/%s", cat)},
			{Name: subcat, Link: from},
		},
		IsProduct: true,
	}

	return templates.Get("admin.html").ExecuteTemplate(w, "base", p)
}

func AdminProductPage(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat)
	err := p.Fetch()
	if err != nil {
		return err
	}

	subs, err := store.GetSubCategories(p.Cat)
	if err != nil {
		return err
	}

	from := fmt.Sprintf("/admin/categories/%s/subcategories/%s/products/%s", p.Cat, p.Subcat, p.Title)
	page := adminPage{
		page: page{
			Admin:   Admin(req),
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

	return templates.Get("admin-product.html").ExecuteTemplate(w, "base", page)
}

func UpdateProduct(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat)
	err := p.Fetch()
	if err != nil {
		return err
	}

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	f, _, err := req.FormFile("Image")
	if f != nil {
		defer f.Close()
	}

	title := req.FormValue("Title")
	desc := req.FormValue("Description")

	p2 := store.NewProduct(title, p.Cat, p.Subcat, store.ProductDescription(desc), store.ProductImage(f))

	if err := p.Update(p2); err != nil {
		return err
	}

	clearEtag(p.Title)

	makeNavbarLinks()
	l := fmt.Sprintf("/admin/categories/%s/subcategories/%s", p.Cat, p2.Subcat)
	w.Header().Set("Location", l)
	w.WriteHeader(http.StatusFound)

	return nil
}

func DeleteProduct(w http.ResponseWriter, req *http.Request) error {
	cat, subcat, vars := getVars(req)

	p := store.NewProduct(vars["title"], cat, subcat)
	if err := p.Fetch(); err != nil {
		return err
	}

	if err := p.Delete(); err != nil {
		return err
	}

	clearEtag(vars["title"])
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
	return templates.Get("confirm.html").ExecuteTemplate(w, "base", p)
}

type wholesaleAdminPage struct {
	page
	Wholesalers []store.User
}

func AdminWholesalers(w http.ResponseWriter, req *http.Request) error {
	users, err := store.GetUsers()
	if err != nil {
		return err
	}

	var wholesalers []store.User
	for _, u := range users {
		if u.Permission == store.Wholesaler {
			wholesalers = append(wholesalers, u)
		}
	}

	p := wholesaleAdminPage{
		page: page{
			Admin: Admin(req),
			Links: getNavbarLinks(req),
			Name:  name,
		},
		Wholesalers: wholesalers,
	}

	return templates.Get("admin-wholesalers.html").ExecuteTemplate(w, "base", p)
}
