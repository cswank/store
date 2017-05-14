package utils

import (
	"fmt"
	"log"

	"github.com/cswank/store/internal/store"
)

func EditBlog() {
	blogs, err := store.Blogs()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("select a blog")
	for i, blog := range blogs {
		fmt.Printf("%d %s\n", i+1, blog.Title)
	}

	var i int
	fmt.Scanf("%d\n", &i)
	key := blogs[i-1]

	blog, err := store.GetBlog(key.ID)
	if err != nil {
		log.Fatal(err)
	}

	blog2, err := store.GetBlog(key.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("new title:")
	var t string

	fmt.Scanf("%q\n", &t)

	blog2.Title = t
	if err := blog.Update(blog2, nil); err != nil {
		log.Fatal("could not update blog title", err)
	}
}

func deleteBlog(blog store.BlogKey) {

}
