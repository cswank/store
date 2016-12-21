package utils

import (
	"fmt"
	"log"

	"github.com/cswank/store/internal/store"
)

//AddUser adds a user from the command line
func EditCategory() {
	cats, err := store.GetCategories()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("select a category")
	for i, cat := range cats {
		fmt.Printf("%d %s\n", i+1, cat)
	}

	var i int
	fmt.Scanf("%d\n", &i)
	cat := cats[i-1]

	fmt.Println("(e)dit or (d)delete?")
	var a string
	fmt.Scanf("%s\n", &a)

	if a == "e" {
		editCatetory(cat)
	} else if a == "d" {
		deleteCategory(cat)
	}
}

func editCatetory(cat string) {
	fmt.Println("(r)ename or (l)list subcategories?")
	var a string
	fmt.Scanf("%s\n", &a)

	if a == "r" {
		renameCategory(cat)
	} else if a == "l" {
		editSubcategories(cat)
	}
}

func deleteCategory(cat string) {

}

func renameCategory(cat string) {

}

func editSubcategories(cat string) {

	subcats, err := store.GetSubCategories(cat)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("select a subcategory")
	for i, subcat := range subcats {
		fmt.Printf("%d %s\n", i+1, subcat)
	}

	var i int
	fmt.Scanf("%d\n", &i)
	subcat := subcats[i-1]

	fmt.Println("(r)ename or (d)delete?")
	var a string
	fmt.Scanf("%s\n", &a)

	if a == "r" {
		fmt.Print("New name: ")
		var n string
		fmt.Scanf("%q\n", &n)
		store.RenameSubcategory(cat, subcat, n)
	} else if a == "d" {
		if err := store.DeleteSubcategory(cat, subcat); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("success")
		}
	}
}
