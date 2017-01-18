package store

import (
	"encoding/json"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Permission int

const (
	Read Permission = iota
	Wholesaler
	Admin
)

type User struct {
	Email          string     `json:"email"`
	Permission     Permission `json:"permission"`
	Password       string     `json:"password,omitempty"`
	HashedPassword []byte     `json:"hashed_password,omitempty"`
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

func (u *User) Save() error {
	if err := u.savePassword(); err != nil {
		return err
	}

	d, _ := json.Marshal(u)
	return db.Put([]Row{{Key: []byte(u.Email), Val: d, Buckets: [][]byte{[]byte("users")}}})
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
