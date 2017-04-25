package shopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/sha3"
)

/*
POST /admin/discounts.json
{
  "discount": {
    "discount_type": "percentage"
    "value": "50.0",
    "code": "fortunamajor",
    "minimum_order_amount": "100.00"
  }
}
*/

type DiscountCode struct {
	ID           int    `json:"id"`
	DiscountType string `json:"discount_type"`
	Value        string `json:"value"`
	Code         string `json:"code"`
	Min          string `json:"minimum_order_amount"`
}

func NewDiscountCode(storeName string) (int, string, error) {
	m := map[string]DiscountCode{
		"discount": DiscountCode{
			DiscountType: "percentage",
			Value:        "50.0",
			Code:         getCode(storeName),
			Min:          "100.00",
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return -1, "", err
	}

	req, err := http.NewRequest("POST", discountURL, &buf)

	if err != nil {
		return -1, "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return -1, "", fmt.Errorf("could not create discount code: %d", resp.StatusCode)
	}

	var m2 map[string]DiscountCode
	if err := json.NewDecoder(resp.Body).Decode(&m2); err != nil {
		return -1, "", err
	}

	dc, ok := m2["discount"]
	if !ok {
		return -1, "", fmt.Errorf("could not create discount code: %v", m2)
	}

	return dc.ID, dc.Code, nil
}

func getCode(name string) string {
	buf := []byte("name")
	h := make([]byte, 64)
	sha3.ShakeSum256(h, buf)
	return string(h)
}
