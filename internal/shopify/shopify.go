package shopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

/*
{
  "product": {
    "title": "Burton Custom Freestyle 151",
    "body_html": "<strong>Good snowboard!<\/strong>",
    "vendor": "Burton",
    "product_type": "Snowboard",
    "images": [
      {
        "src": "http:\/\/example.com\/rails_logo.gif"
      }
    ]
  }
}
*/
m,n.
var (
	ShopifyAPI    string
	ShopifyJSKey  string
	ShopifyDomain string
	productsURL   string
)

func Init() {
	ShopifyAPI = os.Getenv("SHOPIFY_API")
	ShopifyJSKey = os.Getenv("SHOPIFY_JS_KEY")
	ShopifyDomain = os.Getenv("SHOPIFY_DOMAIN")
	if ShopifyJSKey == "" || ShopifyDomain == "" || ShopifyAPI == "" {
		log.Fatal("you must set SHOPIFY_API and SHOPIFY_DOMAIN  and SHOPIFY_DOMAIN")
	}

	productsURL = fmt.Sprintf("%s/%s", ShopifyAPI, "/admin/products.json")
}

type Product struct {
	ID          int    `json:"id,omitempty"`
	Title       string `json:"title"`
	ProductType string `json:"product_type"`
}

func CreateProduct(title, category string) (int, error) {
	m := map[string]Product{
		"product": {Title: title, ProductType: category},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return -1, err
	}

	resp, err := http.Post(productsURL, "application/json", &buf)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var r map[string]Product

	err = json.NewDecoder(resp.Body).Decode(&r)
	return r["product"].ID, err
}
