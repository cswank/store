package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/cswank/store/internal/email"

	"golang.org/x/crypto/bcrypt"
)

type Permission int

const (
	Read Permission = iota
	Wholesaler
	Admin
)

var (
	chars  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	pwBody = `Dear %s,
Please click on the following link to reset your %s password.

https://%s/login/reset/%s

%s
`
)

type Token struct {
	Email   string    `json:"email"`
	Expires time.Time `json:"expiers"`
}

type Address struct {
	Address  string `schema:"address" json:"address,omitempty"`
	Address2 string `schema:"address2" json:"address2,omitempty"`
	Zip      string `schema:"zip" json:"zip,omitempty"`
	City     string `schema:"city" json:"city,omitempty"`
	State    string `schema:"state" json:"state,omitempty"`
	Country  string `schema:"country" json:"country,omitempty"`
}

type User struct {
	Email string `schema:"email" json:"email"`
	//Wholesale stuff
	StoreName       string     `schema:"store_name" json:"store_name,omitempty"`
	Website         string     `schema:"website" json:"website,omitempty"`
	FirstName       string     `schema:"first_name" json:"first_name,omitempty"`
	LastName        string     `schema:"last_name" json:"last_name,omitempty"`
	Address         Address    `schema:"address" json:"address,omitempty"`
	ShippingAddress Address    `schema:"shipping_address" json:"shipping_address,omitempty"`
	Permission      Permission `json:"permission"`
	Password        string     `schema:"password" json:"-"`
	Password2       string     `schema:"confirm-password" json:"-"`
	HashedPassword  []byte     `json:"hashed_password,omitempty"`

	//They clicked on the verification email link
	Verified bool `json:"verified"`
	//Admin approval as a real wholesaler
	Confirmed bool `json:"confirmed,omitempty"`
}

func GetUsers() ([]User, error) {
	var users []User
	return users, db.GetAll(Query{Buckets: [][]byte{[]byte("users")}}, func(key, val []byte) error {
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
	return db.Get([]Query{{Key: []byte(u.Email), Buckets: [][]byte{[]byte("users")}}}, func(key, val []byte) error {
		return json.Unmarshal(val, &u)
	})
}

func (u *User) Save(moreRows ...Query) error {
	if len(u.HashedPassword) == 0 { //a new user
		if err := u.savePassword(); err != nil {
			return err
		}
	}

	d, err := json.Marshal(u)
	if err != nil {
		return err
	}

	rows := []Query{{Key: []byte(u.Email), Val: d, Buckets: [][]byte{[]byte("users")}}}
	rows = append(rows, moreRows...)

	return db.Put(rows)
}

func (u *User) UpdatePassword() error {
	if err := u.savePassword(); err != nil {
		return err
	}

	return u.Save()
}

func (u *User) savePassword() error {
	if len(u.Password) < 8 {
		return errors.New("password is too short (must be at least 8 characters long)")
	}

	if u.Password != u.Password2 {
		return errors.New("passwords don't match")
	}

	u.hashPassword()
	return nil
}

func (u *User) Delete() error {
	return db.Delete([]Query{{Buckets: [][]byte{[]byte("users")}, Key: []byte(u.Email)}})
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
	u.Password2 = ""
	return err
}

func (u *User) GenerateToken() (string, Query, error) {
	token := randStr(32)
	t := Token{
		Email:   u.Email,
		Expires: time.Now().Add(24 * 7 * time.Hour),
	}

	d, err := json.Marshal(t)
	return token, NewQuery(Key(token), Val(d), Buckets("verifications")), err
}

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func VerifyWholesaler(token string) (User, error) {
	var u User
	var email string

	q := []Query{NewQuery(Key(token), Buckets("verifications"))}
	err := db.Get(q, func(key, val []byte) error {
		var t Token
		err := json.Unmarshal(val, &t)
		if err != nil {
			return err
		}
		if time.Now().Sub(t.Expires) > 0 {
			return fmt.Errorf("expired token for %s", t.Email)
		}
		email = t.Email
		return nil
	})
	if err != nil {
		return u, err
	}

	err = db.Get([]Query{NewQuery(Key(email), Buckets("users"))}, func(key, val []byte) error {
		return json.Unmarshal(val, &u)
	})
	if err != nil {
		return u, err
	}

	err = db.Delete(q)
	if err != nil {
		return u, err
	}

	u.Verified = true
	return u, u.Save()
}

func SendPasswordReset(email string) error {
	u := User{Email: email}
	if err := u.Fetch(); err != nil {
		log.Printf("couldn't fetch user for password reset, email: %s, err: %v\n", email, err)
		return nil
	}

	k := randStr(64)
	t := Token{
		Email:   email,
		Expires: time.Now().Add(24 * time.Hour),
	}

	if err := sendPasswordResetEmail(email, k); err != nil {
		return err
	}

	d, _ := json.Marshal(t)
	return db.Put([]Query{NewQuery(Key(k), Val(d), Buckets("tokens"))})
}

func sendPasswordResetEmail(em, token string) error {
	pwBody := `Dear %s,
Please click on the following link to reset your %s password.

https://%s/login/reset/%s

%s
`
	m := email.Msg{
		Email:   em,
		Subject: fmt.Sprintf("%s password reset request", cfg.Name),
		Body:    fmt.Sprintf(pwBody, em, cfg.Domains[0], cfg.Domains[0], token, cfg.Domains[0]),
	}

	return email.Send(m)
}

func GetUserFromResetToken(key string) (User, error) {
	var u User
	var t Token

	q := []Query{NewQuery(Key(key), Buckets("tokens"))}
	err := db.Get(q, func(k, v []byte) error {
		return json.Unmarshal(v, &t)
	})

	if err != nil {
		return u, err
	}

	if time.Now().Sub(t.Expires) > 0 {
		return u, fmt.Errorf("expired token for %s", t.Email)
	}

	u.Email = t.Email
	return u, u.Fetch()
}
