package store_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("products", func() {

	var ()

	BeforeEach(func() {

	})

	JustBeforeEach(func() {
	})

	AfterEach(func() {
	})

	Context("load", func() {
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
