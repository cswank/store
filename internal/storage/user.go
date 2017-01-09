package storage

import (
	"encoding/json"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Permission int

const (
	Reader Permission = iota
	Wholesaler
	Admin
)

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
	Confirmed   bool   `schema:"confirmed" json:"confirmed,omitempty"`

	Permission     Permission `json:"permission"`
	Password       string     `schema:"password" json:"password,omitempty"`
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
	return compare(u.HashedPassword, []byte(pw)) == nil, nil
}

func (u *User) Save() error {
	return u.save(true)
}

func (u *User) Update() error {
	return u.save(false)
}

func (u *User) save(isNew bool) error {
	u.Confirmed = false
	if err := u.checkIfExists(isNew); err != nil {
		return err
	}

	if err := u.savePassword(); err != nil {
		return err
	}

	d, _ := json.Marshal(u)
	return db.Put([]Row{{Key: []byte(u.Email), Val: d, Buckets: [][]byte{[]byte("users")}}})
}

func (u *User) checkIfExists(isNew bool) error {
	if !isNew {
		return nil
	}

	err := db.Get([]Row{{Key: []byte(u.Email), Buckets: [][]byte{[]byte("users")}}}, func(key, val []byte) error {
		return nil
	})

	if err == ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	return ErrAlreadyExists
}

func (u *User) savePassword() error {
	if len(u.Password) < 8 {
		return errors.New("password is too short (must be at least 8 characters long)")
	}

	return u.hashPassword()
}

func (u *User) hashPassword() error {
	var err error
	u.HashedPassword, err = hash(
		[]byte(u.Password),
		bcrypt.DefaultCost,
	)
	u.Password = ""
	return err
}
