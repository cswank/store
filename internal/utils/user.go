package utils

import (
	"fmt"
	"log"

	"github.com/cswank/store/internal/store"
	"github.com/howeyc/gopass"
)

//AddUser adds a user from the command line
func AddUser() {
	var u store.User
	fmt.Print("email: ")
	fmt.Scanf("%s\n", &u.Email)
	u.Permission = store.Admin

	getPasswd(&u)

	if err := u.Save(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("saved")
}

func getPasswd(u *store.User) {
	fmt.Printf("password: ")
	b1, err := gopass.GetPasswd()
	if err != nil {
		log.Fatal(err)
	}
	p1 := string(b1)
	fmt.Printf("again: ")
	b2, err := gopass.GetPasswd()
	if err != nil {
		log.Fatal(err)
	}
	p2 := string(b2)
	if p1 != p2 {
		log.Fatal("passwords don't match")
	}
	u.Password = p1
}

func EditUser() {
	users, err := store.GetUsers()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("select a user")
	listUsers(users)

	var i int
	fmt.Scanf("%d\n", &i)
	u := users[i-1]

	var d string
	fmt.Printf("Delete user %s?  (y/N)\n ", u.Email)
	fmt.Scanf("%s\n", &d)
	if d == "y" {
		u.Delete()
		return
	}
}

func listUsers(users []store.User) {
	fmt.Println("# username  permission")
	for i, u := range users {
		fmt.Println(i+1, u.Email, u.Permission)
	}
}
