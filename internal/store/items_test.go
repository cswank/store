package store_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/caarlos0/env"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cswank/store/internal/store"
)

var _ = Describe("items", func() {

	var (
		tmp string
		//f   *os.File
		cfg store.Config
	)

	BeforeEach(func() {
		var err error
		tmp, err = ioutil.TempDir("", "")
		Expect(err).To(BeNil())
		os.Setenv("STORE_DATADIR", tmp)
		Expect(env.Parse(&cfg)).To(BeNil())
		db := &bolt.DB{}
		store.Init(cfg, store.DB(db))
	})

	AfterEach(func() {
		os.RemoveAll(tmp)
	})

	Context("load", func() {
		It("succeeds", func() {
			i := store.Items{}
			Expect(i.Load("/tmp/items")).To(BeNil())
			store.SetItems(&i)
			Expect(store.GetCategories()).To(ConsistOf([]string{"Cards", "Note Pads"}))
			m := store.GetCategory("Cards")
			Expect(m).To(HaveLen(2))
			Expect(m["Anniversary"]).To(ConsistOf([]string{"10th"}))
			Expect(m["Thank You"]).To(ConsistOf([]string{"thanks-for-your-support"}))
		})
	})

	Context("import", func() {
		FIt("succeeds", func() {
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
