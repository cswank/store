package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type blogPage struct {
	page
	Blog  store.Blog
	Body  template.HTML
	ID    string
	Blogs []store.BlogKey
}

func Blog(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	var b store.Blog
	var err error

	k, ok := vars["blog"]
	if ok {
		b, err = store.GetBlog(k)
	} else {
		b, err = store.CurrentBlog()
	}
	if err != nil {
		return err
	}

	blogs, err := store.Blogs()
	if err != nil {
		return err
	}

	sort.Slice(blogs, func(i, j int) bool {
		return blogs[i].ID > blogs[j].ID
	})

	body := strings.Replace(b.Body, "\n", "\n<br/>", -1)
	p := blogPage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		Blog:  b,
		ID:    b.Key(),
		Blogs: blogs,
		Body:  template.HTML(body),
	}

	return templates.Get("blogs/blogs.html").ExecuteTemplate(w, "base", p)
}

type blogFormPage struct {
	page
	Action string
	Blog   store.Blog
	URI    string
}

func BlogForm(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	var action string
	var b store.Blog
	if vars["blog"] == "new" {
		action = "/admin/blogs"
	} else {
		action = fmt.Sprintf("/admin/blogs/%s", vars["blog"])
		var err error
		b, err = store.GetBlog(vars["blog"])
		if err != nil {
			return err
		}
	}

	p := blogFormPage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		Action: action,
		Blog:   b,
		URI:    fmt.Sprintf("/admin/blogs/%s", b.Key()),
	}

	return templates.Get("admin/blog-form.html").ExecuteTemplate(w, "base", p)
}

func BlogImage(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	img, err := store.GetBlogImage(vars["blog"])
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
	return nil

}

func ManageBlogs(w http.ResponseWriter, req *http.Request) error {
	blogs, err := store.Blogs()
	fmt.Println("blogs", blogs, err)
	if err != nil {
		return err
	}

	sort.Slice(blogs, func(i, j int) bool {
		return blogs[i].ID > blogs[j].ID
	})

	p := blogPage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		Blogs: blogs,
	}

	return templates.Get("admin/blogs.html").ExecuteTemplate(w, "base", p)
}

func DeleteBlog(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	b, err := store.GetBlog(vars["blog"])
	if err != nil {
		return err
	}

	if err := b.Delete(); err != nil {
		return err
	}

	w.Header().Set("Location", "/admin/blogs")
	w.WriteHeader(http.StatusFound)
	return nil
}

func UpdateBlog(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	b, err := store.GetBlog(vars["blog"])

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	ff, _, err := req.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		return err
	} else if err == nil {
		defer ff.Close()
	}

	var b2 store.Blog
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&b2, req.PostForm); err != nil {
		return err
	}

	if err := b.Update(b2, ff); err != nil {
		return err
	}

	w.Header().Set("Location", "/admin/blogs")
	w.WriteHeader(http.StatusFound)
	return nil
}

func CreateBlog(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	ff, _, err := req.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		return err
	} else if err == nil {
		defer ff.Close()
	}

	var b store.Blog
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&b, req.PostForm); err != nil {
		return err
	}

	w.Header().Set("Location", "/admin/blogs")
	w.WriteHeader(http.StatusFound)
	return b.Save(ff)
}
