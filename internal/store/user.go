package store

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Permission int

const (
	Read Permission = iota
	Wholesaler
	Admin
)

var (
	chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

type Verification struct {
	Email   string    `json:"email"`
	Expires time.Time `json:"expiers"`
}

type User struct {
	Email string `schema:"email" json:"email"`
	//Wholesale stuff
	FirstName   string `schema:"first_name" json:"first_name,omitempty"`
	LastName    string `schema:"last_name" json:"last_name,omitempty"`
	CompanyName string `schema:"company_name" json:"company_name,omitempty"`
	Address     string `schema:"address" json:"address,omitempty"`
	Address2    string `schema:"address2" json:"address2,omitempty"`
	Zip         string `schema:"zip" json:"zip,omitempty"`
	City        string `schema:"city" json:"city,omitempty"`
	State       string `schema:"state" json:"state,omitempty"`
	Country     string `schema:"country" json:"country,omitempty"`

	Permission     Permission `json:"permission"`
	Password       string     `schema:"password" json:"password,omitempty"`
	HashedPassword []byte     `json:"hashed_password,omitempty"`

	//They clicked on the verification email link
	Verified bool `json:"verified"`
	//Admin approval as a real wholesaler
	Confirmed bool `json:"confirmed,omitempty"`
}

func GetUsers() ([]User, error) {
	var users []User
	return users, db.GetAll(Row{Buckets: [][]byte{[]byte("users")}}, func(key, val []byte) error {
		var u User
		if err := json.Unmarshal(val, &u); err != nil {
			return err
		}
		u.HashedPassword = []byte{}
		users = append(users, u)
		return nil
	})
}

func (u *User) Fetch() error {
	return db.Get([]Row{{Key: []byte(u.Email), Buckets: [][]byte{[]byte("users")}}}, func(key, val []byte) error {
		return json.Unmarshal(val, &u)
	})
}

func (u *User) Save(opts ...func() Row) error {

	if err := u.savePassword(); err != nil {
		return err
	}

	d, err := json.Marshal(u)
	if err != nil {
		return err
	}

	rows := []Row{{Key: []byte(u.Email), Val: d, Buckets: [][]byte{[]byte("users")}}}
	for _, o := range opts {
		rows = append(rows, o())
	}

	return db.Put(rows)
}

func (u *User) savePassword() error {
	if len(u.Password) < 8 {
		return errors.New("password is too short (must be at least 8 characters long)")
	}

	u.hashPassword()
	return nil
}

func (u *User) Delete() error {
	return db.Delete([]Row{{Buckets: [][]byte{[]byte("users")}, Key: []byte(u.Email)}})
}

func (u *User) CheckPassword() (bool, error) {
	pw := u.Password
	if len(u.HashedPassword) == 0 {
		if err := u.Fetch(); err != nil {
			return false, err
		}
	}
	return bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(pw)) == nil, nil
}

func (u *User) hashPassword() error {
	var err error
	u.HashedPassword, err = bcrypt.GenerateFromPassword(
		[]byte(u.Password),
		bcrypt.DefaultCost,
	)
	u.Password = ""
	return err
}

func (u *User) GenerateToken() (string, Row, error) {
	token := randStr(32)
	v := Verification{
		Email:   u.Email,
		Expires: time.Now().Add(24 * 7 * time.Hour),
	}

	d, err := json.Marshal(v)
	return token, NewRow(Key(token), Val(d), Buckets("verifications")), err
}

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
