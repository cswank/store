package handlers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type blogPage struct {
	page
	Blog  store.Blog
	ID    string
	Blogs []store.BlogKey
}

func Blog(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	var b *store.Blog
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

	p := blogPage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
		},
		Blog:  *b,
		ID:    b.Key(),
		Blogs: blogs,
	}

	return templates.Get("blogs/blogs.html").ExecuteTemplate(w, "base", p)
}

type blogFormPage struct {
	page
	Action string
	Blog   store.Blog
}

func BlogForm(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	var action string
	var b store.Blog
	if vars["blog"] == "new" {
		action = "/admin/blogs"
	} else {
		action = fmt.Sprintf("/admin/blogs/%s", vars["blog"])
	}

	p := blogFormPage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
		},
		Action: action,
		Blog:   b,
	}

	return templates.Get("blogs/form.html").ExecuteTemplate(w, "base", p)
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

func CreateBlog(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	ff, _, err := req.FormFile("image")
	fmt.Println("blog image", ff)
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

	return b.Save(ff)
}
