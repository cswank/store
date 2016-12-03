package store

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

type Permission int

const (
	Read Permission = iota
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
	return users, db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u User
			if err := json.Unmarshal(v, &u); err != nil {
				return err
			}
			u.HashedPassword = []byte{}
			users = append(users, u)
		}
		return nil
	})
}

func (u *User) Fetch() error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		v := b.Get([]byte(u.Email))
		if len(v) == 0 {
			return ErrNotFound
		}
		return json.Unmarshal(v, u)
	})
}

func (u *User) Save() error {
	if err := u.savePassword(); err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		d, _ := json.Marshal(u)
		b := tx.Bucket([]byte("users"))
		return b.Put([]byte(u.Email), d)
	})
}

func (u *User) savePassword() error {
	if len(u.Password) < 8 {
		return errors.New("password is too short (must be at least 8 characters long)")
	}

	u.hashPassword()
	return nil
}

func (u *User) Delete() error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		return b.Delete([]byte(u.Email))
	})
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
