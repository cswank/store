package store_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/shopify"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/store/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("products", func() {

	var (
		f       *os.File
		db      *mock.DB
		id      int
		ts      *httptest.Server
		cfg     config.Config
		buckets map[string][]mock.Result
		errs    []error
	)

	BeforeEach(func() {
		buckets = map[string][]mock.Result{}
		errs = []error{}
		var err error
		f, err = os.Open("testdata/product.png")
		Expect(err).To(BeNil())

		id = 1
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				var m map[string]shopify.Product
				Expect(json.NewDecoder(r.Body).Decode(&m)).To(BeNil())
				p := m["product"]
				p.ID = id
				p.Variants = []shopify.Variant{
					{ID: id},
				}
				m["product"] = p
				id++
				Expect(json.NewEncoder(w).Encode(m)).To(BeNil())
			}
		}))

		os.Setenv("SHOPIFY_DOMAIN", ts.URL)
		os.Setenv("SHOPIFY_API", ts.URL)
		os.Setenv("SHOPIFY_JS_KEY", "xyz")
		Expect(env.Parse(&cfg)).To(BeNil())
	})

	JustBeforeEach(func() {
		db = mock.NewDB(
			buckets,
			errs,
		)
		shopify.Init(cfg)
		store.Init(cfg, store.SetDB(db))
	})

	AfterEach(func() {
		f.Close()
	})

	Context("GetCategories", func() {

		Context("success", func() {

			BeforeEach(func() {
				buckets = map[string][]mock.Result{
					"products": []mock.Result{{Key: []byte("Cards")}, {Key: []byte("Note Pads")}},
				}
				errs = []error{
					nil,
					nil,
				}
			})

			It("succeeds", func() {
				cats, err := store.GetCategories()
				Expect(err).To(BeNil())
				Expect(cats).To(ConsistOf([]string{"Cards", "Note Pads"}))
			})
		})
	})

	Context("GetSubCategories", func() {

		Context("success", func() {

			BeforeEach(func() {
				buckets = map[string][]mock.Result{
					"products Cards": []mock.Result{{Key: []byte("you-are-fucked")}, {Key: []byte("fortune-cookie")}},
				}
				errs = []error{
					nil,
					nil,
				}
			})

			It("succeeds", func() {
				cats, err := store.GetSubCategories("Cards")
				Expect(err).To(BeNil())
				Expect(cats).To(ConsistOf([]string{"you-are-fucked", "fortune-cookie"}))
			})
		})
	})

	Describe("Product", func() {

		var (
			prod *store.Product
		)

		BeforeEach(func() {
			prod = store.NewProduct("you-are-fucked", "Cards", "Happy Birthday", store.ProductDescription("Blah blah blah!"))
		})

		Describe("Delete", func() {

			Context("valid product", func() {

				BeforeEach(func() {
					buckets = map[string][]mock.Result{
						"products Cards Happy Birthday": []mock.Result{{Key: []byte("you-are-fucked"), Val: []byte(`{"id": "33", "description": "blah"}`)}},
					}
					errs = []error{
						nil,
						nil,
					}

				})

				It("succeeds", func() {
					Expect(prod.Delete()).To(BeNil())
					Expect(db.Rows).To(HaveLen(2))

					//query
					r := db.Rows[0]
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("products"))
					Expect(string(r.Buckets[1])).To(Equal("Cards"))
					Expect(string(r.Buckets[2])).To(Equal("Happy Birthday"))
					Expect(string(r.Key)).To(Equal("you-are-fucked"))

					r = db.Rows[1]
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("images"))
					Expect(string(r.Buckets[1])).To(Equal("products"))
					Expect(string(r.Buckets[2])).To(Equal("you-are-fucked"))
				})
			})
		})

		Describe("Update", func() {

			Context("moving to a different subcatetory", func() {

				BeforeEach(func() {
					buckets = map[string][]mock.Result{
						"products Cards Happy Birthday": []mock.Result{{Key: []byte("you-are-fucked"), Val: []byte(`{"id": "33", "description": "blah"}`)}},
					}
					errs = []error{
						nil,
						nil,
						nil,
						nil,
					}

				})

				It("succeeds", func() {
					p2 := store.NewProduct(prod.Title, prod.Cat, "Anniversary", store.ProductDescription("Blah blah blah!"))
					Expect(prod.Update(p2)).To(BeNil())
					Expect(db.Rows).To(HaveLen(4))

					r := db.Rows[0]
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("products"))
					Expect(string(r.Buckets[1])).To(Equal("Cards"))
					Expect(string(r.Buckets[2])).To(Equal("Happy Birthday"))
					Expect(string(r.Key)).To(Equal("you-are-fucked"))

					r = db.Rows[1]
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("products"))
					Expect(string(r.Buckets[1])).To(Equal("Cards"))
					Expect(string(r.Buckets[2])).To(Equal("Happy Birthday"))
					Expect(string(r.Key)).To(Equal("you-are-fucked"))

					r = db.Rows[2]
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("products"))
					Expect(string(r.Buckets[1])).To(Equal("Cards"))
					Expect(string(r.Buckets[2])).To(Equal("Anniversary"))
					Expect(string(r.Key)).To(Equal("you-are-fucked"))
					Expect(string(r.Val)).To(MatchJSON(`{"description":"blah","id":"33"}`))
				})
			})
		})

		Describe("Add", func() {

			Context("valid insert", func() {

				BeforeEach(func() {
					buckets = map[string][]mock.Result{
						"products Cards Happy Birthday": []mock.Result{{Key: []byte("fortune-cookie")}},
					}
					errs = []error{
						nil,
						nil,
					}

				})

				It("succeeds", func() {
					Expect(prod.Add(f)).To(BeNil())
					Expect(db.Rows).To(HaveLen(4))

					//query
					r := db.Rows[0]
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("products"))
					Expect(string(r.Buckets[1])).To(Equal("Cards"))
					Expect(string(r.Buckets[2])).To(Equal("Happy Birthday"))

					//insert image
					r = db.Rows[1]
					Expect(string(r.Key)).To(Equal("image.png"))
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("images"))
					Expect(string(r.Buckets[1])).To(Equal("products"))
					Expect(string(r.Buckets[2])).To(Equal("you-are-fucked"))

					r = db.Rows[2]
					Expect(string(r.Key)).To(Equal("thumb.png"))
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("images"))
					Expect(string(r.Buckets[1])).To(Equal("products"))
					Expect(string(r.Buckets[2])).To(Equal("you-are-fucked"))

					r = db.Rows[3]
					Expect(string(r.Key)).To(Equal("you-are-fucked"))
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("products"))
					Expect(string(r.Buckets[1])).To(Equal("Cards"))
					Expect(string(r.Buckets[2])).To(Equal("Happy Birthday"))
					Expect(string(r.Val)).To(MatchJSON(`{"description":"Blah blah blah!","id":"1"}`))
				})
			})

			Context("duplicate insert", func() {

				BeforeEach(func() {
					buckets = map[string][]mock.Result{
						"products Cards Happy Birthday": []mock.Result{{Key: []byte("fortune-cookie")}, {Key: []byte("you-are-fucked")}},
					}
					errs = []error{
						nil,
						nil,
					}
				})

				It("bombs out", func() {
					Expect(prod.Add(f)).To(MatchError("product already exists"))
					Expect(db.Rows).To(HaveLen(1))
					//query
					r := db.Rows[0]
					Expect(r.Buckets).To(HaveLen(3))
					Expect(string(r.Buckets[0])).To(Equal("products"))
					Expect(string(r.Buckets[1])).To(Equal("Cards"))
					Expect(string(r.Buckets[2])).To(Equal("Happy Birthday"))
				})
			})
		})
	})
})

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
