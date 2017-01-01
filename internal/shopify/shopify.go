package shopify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cswank/store/internal/config"
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

var (
	productsURL string
	imagesURL   string
	variantsURL string
	deleteURL   string

	cfg config.Config
)

func Init(c config.Config) {
	cfg = c

	productsURL = fmt.Sprintf("%s/%s", cfg.ShopifyAPI, "admin/products.json")
	imagesURL = fmt.Sprintf("%s/%s", cfg.ShopifyAPI, "/admin/products/%s/images.json")
	variantsURL = fmt.Sprintf("%s/%s", cfg.ShopifyAPI, "admin/variants/%d.json")
	deleteURL = fmt.Sprintf("%s/%s", cfg.ShopifyAPI, "admin/products/%s.json")
}

type Img struct {
	Src string `json:"src,omitempty"`
}

type Variant struct {
	ID    int    `json:"id"`
	Price string `json:"price"`
}

type Product struct {
	ID             int       `json:"id,omitempty"`
	Title          string    `json:"title"`
	ProductType    string    `json:"product_type"`
	PublishedAt    string    `json:"published_at"`
	PublishedScope string    `json:"published_scope"`
	Images         []Img     `json:"images"`
	Variants       []Variant `json:"variants"`
}

func Create(title, category, price string) (string, error) {
	m := map[string]Product{
		"product": {
			Title:          title,
			ProductType:    category,
			PublishedAt:    time.Now().Format("2006-01-02T15:04:05-0700"),
			PublishedScope: "global",
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return "", err
	}

	resp, err := http.Post(productsURL, "application/json", &buf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var r map[string]Product

	err = json.NewDecoder(resp.Body).Decode(&r)

	p, ok := r["product"]
	if !ok {
		return "", errors.New("unexpected response from shopify")
	}

	if len(p.Variants) != 1 {
		return "", errors.New("unexpected response from shopify")
	}

	v := p.Variants[0]

	if err := updatePrice(v.ID, price); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", p.ID), nil
}

func Delete(id string) error {
	uri := fmt.Sprintf(deleteURL, id)

	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}

func AddImage(id string, img []byte) error {
	s := base64.StdEncoding.EncodeToString(img)
	buf := bytes.NewBufferString(fmt.Sprintf(`{"image": {"attachment": "%s"}}`, s))
	u := fmt.Sprintf(imagesURL, id)
	resp, err := http.Post(u, "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status when adding image to spotify: %d", resp.StatusCode)
	}

	return nil
}

func updatePrice(id int, price string) error {
	m := map[string]Variant{
		"variant": {
			ID:    id,
			Price: price,
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf(variantsURL, id), &buf)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
