package store_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/caarlos0/env"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cswank/store/internal/shopify"
	"github.com/cswank/store/internal/store"
	"github.com/cswank/store/internal/store/mock"
)

var _ = Describe("items", func() {

	var (
		tmp   string
		cfg   store.Config
		items string
		ts    *httptest.Server
		id    int
		errs  []error
		vals  [][]byte
		db    *mock.Store
	)

	BeforeEach(func() {
		id = 1
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var m map[string]shopify.Product
			Expect(json.NewDecoder(r.Body).Decode(&m)).To(BeNil())
			p := m["product"]
			p.ID = id
			m["product"] = p
			id++
			Expect(json.NewEncoder(w).Encode(m)).To(BeNil())
		}))

		var err error
		tmp, err = ioutil.TempDir("", "")
		Expect(err).To(BeNil())
		os.Setenv("STORE_DATADIR", tmp)
		os.Setenv("SHOPIFY_DOMAIN", ts.URL)
		os.Setenv("SHOPIFY_APIKEY", "xyz")

		Expect(env.Parse(&cfg)).To(BeNil())

		cmd := "unzip"
		args := []string{"./testdata/items.zip", "-d", tmp}
		Expect(exec.Command(cmd, args...).Run()).To(BeNil())
		items = filepath.Join(tmp, "items")
	})

	JustBeforeEach(func() {
		db = mock.NewStore(vals, errs)
		shopify.Init()
		store.Init(cfg, store.SetDB(db))
	})

	AfterEach(func() {
		ts.Close()
		os.RemoveAll(tmp)
	})

	Context("load", func() {

		Context("all new", func() {

			BeforeEach(func() {
				errs = []error{
					store.ErrNotFound,
					nil,
					store.ErrNotFound,
					nil,
				}

				vals = [][]byte{
					nil,
					nil,
					nil,
					nil,
				}
			})

			FIt("succeeds", func() {
				Expect(store.GetCategories()).To(ConsistOf([]string{"Cards"}))
				m := store.GetCategory("Cards")
				Expect(m).To(HaveLen(2))
				Expect(m["Anniversary"]).To(ConsistOf([]string{"10th"}))
				Expect(m["Thank You"]).To(ConsistOf([]string{"thanks-for-your-support"}))

				for _, img := range []string{"image.jpg", "thumb.jpg"} {
					pth := filepath.Join(items, "Cards", "Thank You", "thanks-for-your-support", img)
					Expect(pth).Should(BeARegularFile())
				}

				Expect(id).To(Equal(3))
				Expect(db.Puts["10th-products"]).To(Equal("1"))
				Expect(db.Puts["thanks-for-your-support-products"]).To(Equal("2"))
			})

		})

		Context("already exists", func() {

			BeforeEach(func() {
				errs = []error{
					nil,
					nil,
				}

				vals = [][]byte{
					nil,
					nil,
				}
			})

			FIt("succeeds", func() {
				Expect(store.GetCategories()).To(ConsistOf([]string{"Cards"}))
				m := store.GetCategory("Cards")
				Expect(m).To(HaveLen(2))
				Expect(m["Anniversary"]).To(ConsistOf([]string{"10th"}))
				Expect(m["Thank You"]).To(ConsistOf([]string{"thanks-for-your-support"}))

				for _, img := range []string{"image.jpg", "thumb.jpg"} {
					pth := filepath.Join(items, "Cards", "Thank You", "thanks-for-your-support", img)
					Expect(pth).Should(BeARegularFile())
				}

				Expect(id).To(Equal(1)) //shopify api wasn't called
				Expect(db.Puts).To(HaveLen(0))
			})

		})
	})

	Context("import", func() {
		It("succeeds", func() {
			f, err := os.Open("/tmp/items.zip")
			Expect(err).To(BeNil())
			Expect(store.ImportItems(f)).To(BeNil())
			f.Close()
			z := filepath.Join(tmp, "archive", "items-0.zip")
			Expect(fileExists(z)).To(BeTrue())

			Expect(store.GetCategories()).To(ConsistOf([]string{"Cards", "Note Pads"}))
			m := store.GetCategory("Cards")
			Expect(m).To(HaveLen(2))
			Expect(m["Anniversary"]).To(ConsistOf([]string{"10th"}))
			Expect(m["Thank You"]).To(ConsistOf([]string{"thanks-for-your-support"}))
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
