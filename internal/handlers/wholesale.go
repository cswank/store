package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cswank/store/internal/email"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/templates"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type formPage struct {
	page
	Captcha        bool
	CaptchaSiteKey string
	ShowMessage    bool
}

func Purchase(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func Wholesale(w http.ResponseWriter, req *http.Request) error {
	if Wholesaler(req) {
		return getWholesaleForm(w, req)
	} else if NewWholesaler(req) {
		return getWholesaleProcessing(w, req)
	}
	w.Header().Set("Location", "/login?from=/wholesale")
	w.WriteHeader(http.StatusFound)
	return nil
}

type wholesalePage struct {
	page
	Products map[string]map[string][]product
	Items    map[string]product
}

func getWholesaleForm(w http.ResponseWriter, req *http.Request) error {
	cats, err := store.GetCategories()
	if err != nil {
		return err
	}

	prods, items, err := getWholesaleProducts(cats)
	if err != nil {
		return err
	}

	p := wholesalePage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		Products: prods,
		Items:    items,
	}

	return templates.Get("wholesale/form.html").ExecuteTemplate(w, "base", p)
}

//the wholesaler has not clicked on the verify link in the email
//that was sent, or the admin has not clicked the confirm button
func getWholesaleProcessing(w http.ResponseWriter, req *http.Request) error {
	p := page{
		Links:   getNavbarLinks(req),
		Name:    cfg.Name,
		Head:    html["head"],
		Message: "Your wholesale application is still being processed.  You will receive an additional email once your application is approved and you will then be able to purchase items at wholesale prices",
	}
	return templates.Get("wholesale/pending.html").ExecuteTemplate(w, "base", p)
}

func Invoice(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	if req.Method == "GET" {
		return previewInvoice(w, req)
	}

	return sendInvoice(w, req)
}

type invoiceEmail struct {
	Number     int
	Date       time.Time
	StyleSheet string
	Total      float64
	Price      string
	Products   []invoiceProduct
	Customer   *store.User
}

type invoiceSent struct {
	page
	Customer *store.User
}

func sendInvoice(w http.ResponseWriter, req *http.Request) error {
	products, total := getInvoiceProducts(req)
	u := getUser(req)

	i := invoiceEmail{
		Number:     0,
		Date:       time.Now(),
		Total:      total,
		StyleSheet: cfg.InvoiceStylesheet,
		Customer:   u,
		Products:   products,
		Price:      cfg.DefaultPrice,
	}

	var buf bytes.Buffer
	if err := templates.Get("wholesale/invoice.html").ExecuteTemplate(&buf, "invoice", i); err != nil {
		return err
	}

	msg := email.Msg{
		To:      u.Email,
		From:    cfg.Email,
		Subject: fmt.Sprintf("Invoice #%d from %s", i.Number, cfg.Domains[0]),
		Body:    buf.String(),
	}

	if err := email.Send(msg); err != nil {
		return err
	}

	p := invoiceSent{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		Customer: u,
	}

	return templates.Get("wholesale/invoice-sent.html").ExecuteTemplate(w, "base", p)
}

type invoiceProduct struct {
	Title    string
	Total    string
	Quantity int
}

type invoicePreview struct {
	page
	Products []invoiceProduct
	Price    string
	Total    string
}

func previewInvoice(w http.ResponseWriter, req *http.Request) error {
	products, total := getInvoiceProducts(req)

	p := invoicePreview{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		Products: products,
		Price:    cfg.DefaultPrice,
		Total:    fmt.Sprintf("%.02f", total),
	}

	return templates.Get("wholesale/preview.html").ExecuteTemplate(w, "base", p)
}

func getInvoiceProducts(req *http.Request) ([]invoiceProduct, float64) {
	var products []invoiceProduct
	price, _ := strconv.ParseFloat(cfg.DefaultPrice, 32)
	var total float64
	for key, values := range req.Form { // range over map
		for _, value := range values { // range over []string
			if value == "0" {
				continue
			}
			q, err := strconv.ParseFloat(value, 32)
			if err != nil {
				log.Println("couldn't parse form value", key, value, err)
				continue
			}
			t := price * q
			total += t
			products = append(products, invoiceProduct{
				Total:    fmt.Sprintf("%.02f", t),
				Title:    key,
				Quantity: int(q),
			})
		}
	}
	return products, total
}

func ConfirmInvoice(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func WholesaleApplication(w http.ResponseWriter, req *http.Request) error {
	params := req.URL.Query()
	p := formPage{
		page: page{
			Links: getNavbarLinks(req),
			Name:  cfg.Name,
			Head:  html["head"],
		},
		ShowMessage:    params.Get("success") != "",
		CaptchaSiteKey: cfg.RecaptchaSiteKey,
		Captcha:        true,
	}

	return templates.Get("wholesale/application-form.html").ExecuteTemplate(w, "base", p)
}

func getWholesaleProducts(cats []string) (map[string]map[string][]product, map[string]product, error) {
	m := map[string]map[string][]product{}
	m2 := map[string]product{}
	for _, cat := range cats {

		price, err := store.GetPrice(cat)
		if err != nil {
			return nil, nil, err
		}

		subcats, err := store.GetSubCategories(cat)
		if err != nil {
			return nil, nil, err
		}
		a := map[string][]product{}
		for _, subcat := range subcats {
			prods, err := store.GetProducts(cat, subcat)
			if err != nil {
				return nil, nil, err
			}
			pp := getProducts(cat, subcat, prods, price.WholesalePrice)
			a[subcat] = pp
			for _, p := range pp {
				m2[p.ID] = p
			}
		}
		m[cat] = a
	}
	return m, m2, nil
}

func WholesaleThanks(w http.ResponseWriter, req *http.Request) error {
	msg := fmt.Sprintf("Your email address is confirmed.  As soon as the site administrator approves your account you will be able to log into %s and make purchases at wholesale prices", cfg.Domains[0])
	p := page{
		Links:   getNavbarLinks(req),
		Name:    cfg.Name,
		Head:    html["head"],
		Message: msg,
	}

	return templates.Get("wholesale/thanks.html").ExecuteTemplate(w, "base", p)
}

func WholesaleApply(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	u := store.User{
		Permission: store.Wholesaler,
	}

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&u, req.PostForm); err != nil {
		return err
	}

	token, row, err := u.GenerateToken()
	if err != nil {
		return err
	}

	msg := email.Msg{
		To:      u.Email,
		From:    cfg.Email,
		Subject: fmt.Sprintf("Thank you for applying to %s", cfg.Domains[0]),
		Body:    getWholesaleVerificationBody(u, token),
	}

	if err := email.Send(msg); err != nil {
		return err
	}

	msg = email.Msg{
		To:      cfg.Email,
		From:    cfg.Email,
		Subject: fmt.Sprintf("New wholesaler application for %s", cfg.Domains[0]),
		Body: fmt.Sprintf(
			"%s has applied as a wholesaler.  Click on %s to approve the application.",
			u.Email,
			fmt.Sprintf("https://%s/admin/wholesalers/%s", cfg.Domains[0], u.Email),
		),
	}

	if err := email.Send(msg); err != nil {
		return err
	}

	if err := u.Save(row); err != nil {
		return err
	}

	w.Header().Set("Location", "/wholesale/thanks")
	w.WriteHeader(http.StatusFound)
	return nil
}

func getWholesaleVerificationBody(u store.User, token string) string {

	tmpl := `Hello %s,
Thank you for applying at %s as a wholesaler.  Please
click on this link in order to verify your email address.

https://%s/wholesale/application/%s

As soon as the site administrator approves your application you will
receive an additional email informing you that you have been approved.
You will then be able to log into %s to purchase our products.

Thanks!

%s`

	return fmt.Sprintf(tmpl, u.FirstName, cfg.Domains[0], cfg.Domains[0], token, cfg.Domains[0], cfg.Email)
}

func WholesaleVerify(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	p := page{
		Links:   getNavbarLinks(req),
		Admin:   Admin(req),
		Shopify: shopifyKey,
		Head:    html["head"],
		Name:    name,
	}
	var f func(io.Writer, string, interface{}) error
	u, err := store.VerifyWholesaler(vars["token"])

	if err != nil {
		log.Printf("failed to confirm user %s with token %s, err: %v\n", u.Email, vars["token"], err)
		f = templates.Get("wholesale/pending.html").ExecuteTemplate
		p.Message = "We were unable to confirm your email address.  If you applied more than 7 days ago your application has expired and you will have to re-apply.  Sorry for the inconvenience."
	} else if u.Verified && u.Confirmed {
		f = templates.Get("wholesale/welcome.html").ExecuteTemplate
	} else {
		f = templates.Get("wholesale/pending.html").ExecuteTemplate
		p.Message = "Thank you.  Your email address has been confirmed. Once the site administrator approves your application you will receive an email from us."
	}

	return f(w, "base", p)
}
