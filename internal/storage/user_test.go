package storage_test

import (
	"github.com/cswank/store/internal/config"
	"github.com/cswank/store/internal/storage"
	"github.com/cswank/store/internal/storage/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("users", func() {

	var (
		db      *mock.DB
		cfg     config.Config
		buckets map[string][]mock.Result
		errs    []error
		user    storage.User
		hash    func([]byte, int) ([]byte, error)
		cmp     func([]byte, []byte) error
	)

	BeforeEach(func() {
		hash = func([]byte, int) ([]byte, error) { return []byte("bytez"), nil }
		cmp = func([]byte, []byte) error { return nil }
		buckets = map[string][]mock.Result{}
		errs = []error{}
		user = storage.User{
			Email:      "sal@trinkets.com",
			Password:   "aaabbbccc",
			Permission: storage.Wholesaler,
		}
	})

	JustBeforeEach(func() {
		db = mock.NewDB(buckets, errs)
		storage.Init(cfg, storage.DB(db), storage.Compare(cmp), storage.Hash(hash))
	})

	AfterEach(func() {
	})

	Context("Save", func() {

		Context("success", func() {

			BeforeEach(func() {
				errs = []error{
					storage.ErrNotFound,
					nil,
				}
			})

			It("succeeds", func() {
				Expect(user.Save()).To(BeNil())
				Expect(db.Rows).To(HaveLen(2))
				r := db.Rows[0]
				Expect(r.Buckets).To(HaveLen(1))
				Expect(string(r.Buckets[0])).To(Equal("users"))
				Expect(string(r.Key)).To(Equal("sal@trinkets.com"))

				r = db.Rows[1]
				Expect(r.Buckets).To(HaveLen(1))
				Expect(string(r.Buckets[0])).To(Equal("users"))
				Expect(string(r.Key)).To(Equal("sal@trinkets.com"))
				Expect(string(r.Val)).To(Equal(`{"email":"sal@trinkets.com","permission":1,"hashed_password":"Ynl0ZXo="}`))
			})
		})

		Context("user already exists", func() {

			BeforeEach(func() {
				errs = []error{
					storage.ErrAlreadyExists,
					nil,
				}
			})

			It("succeeds", func() {
				Expect(user.Save()).To(MatchError("already exists"))
			})
		})
	})

	Context("Update", func() {

		Context("success", func() {

			BeforeEach(func() {
				buckets = map[string][]mock.Result{
					"users": []mock.Result{{Key: []byte("sal@trinkets.com")}, {Key: []byte("Note Pads")}},
				}
				errs = []error{
					nil,
					nil,
				}
			})

			It("succeeds", func() {
				Expect(user.Update()).To(BeNil())
				Expect(db.Rows).To(HaveLen(1))
				r := db.Rows[0]
				Expect(r.Buckets).To(HaveLen(1))
				Expect(string(r.Buckets[0])).To(Equal("users"))
				Expect(string(r.Key)).To(Equal("sal@trinkets.com"))
				Expect(string(r.Val)).To(Equal(`{"email":"sal@trinkets.com","permission":1,"hashed_password":"Ynl0ZXo="}`))
			})
		})
	})

	Context("Delete", func() {

		Context("success", func() {

			BeforeEach(func() {
				errs = []error{
					nil,
				}
			})

			It("succeeds", func() {
				Expect(user.Delete()).To(BeNil())
				Expect(db.Rows).To(HaveLen(1))
				r := db.Rows[0]
				Expect(r.Buckets).To(HaveLen(1))
				Expect(string(r.Buckets[0])).To(Equal("users"))
				Expect(string(r.Key)).To(Equal("sal@trinkets.com"))
				Expect(string(r.Val)).To(Equal(``))
			})
		})
	})

	Context("Fetch", func() {

		Context("success", func() {

			BeforeEach(func() {
				user = storage.User{Email: "sal@trinkets.com"}
				buckets = map[string][]mock.Result{
					"users": []mock.Result{{Key: []byte("sal@trinkets.com"), Val: []byte(`{"email":"sal@trinkets.com","permission":1,"hashed_password":"Ynl0ZXo=","address":"123 4 Street"}`)}},
				}

				errs = []error{
					nil,
				}
			})

			It("succeeds", func() {
				Expect(user.Fetch()).To(BeNil())
				Expect(user.Address).To(Equal("123 4 Street"))
				Expect(db.Rows).To(HaveLen(1))
				r := db.Rows[0]
				Expect(r.Buckets).To(HaveLen(1))
				Expect(string(r.Buckets[0])).To(Equal("users"))
				Expect(string(r.Key)).To(Equal("sal@trinkets.com"))
				Expect(string(r.Val)).To(Equal(``))
			})
		})
	})
})
