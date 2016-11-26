package handlers

import (
	"fmt"
	"net/http"
)

func DeleteCategory(w http.ResponseWriter, req *http.Request) {
	fmt.Println("delete category")
}
