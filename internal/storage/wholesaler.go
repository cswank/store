package storage

import "encoding/json"

type Wholesaler struct {
	User
	CompanyName string `schema:"company_name" json:"company_name"`
	Address     string `schema:"address" json:"address"`
	Address2    string `schema:"address2" json:"address2"`
	Zip         string `schema:"zip" json:"zip"`
	City        string `schema:"city" json:"city"`
	State       string `schema:"state" json:"state"`
	Country     string `schema:"country" json:"country"`
	Confirmed   bool   `schema:"confirmed" json:"confirmed"`
}

func GetWholesalers() ([]Wholesaler, error) {
	var wholesalers []Wholesaler
	return wholesalers, db.GetAll(Row{Buckets: [][]byte{[]byte("wholesalers")}}, func(key, val []byte) error {
		var w Wholesaler
		if err := json.Unmarshal(val, &w); err != nil {
			return err
		}
		w.User.HashedPassword = []byte{}
		wholesalers = append(wholesalers, w)
		return nil
	})
}

func (w *Wholesaler) Save(isNew bool) error {
	w.Permission = Wholesale
	if err := w.checkIfExists(isNew); err != nil {
		return err
	}

	if err := w.User.savePassword(); err != nil {
		return err
	}

	d, err := json.Marshal(w)
	if err != nil {
		return err
	}

	return db.Put([]Row{{Key: []byte(w.Email), Val: d, Buckets: [][]byte{[]byte("wholesalers")}}})
}

func (w *Wholesaler) checkIfExists(isNew bool) error {
	if !isNew {
		return nil
	}

	w.Confirmed = false

	var tmp Wholesaler
	err := db.Get([]Row{{Key: []byte(w.User.Email), Buckets: [][]byte{[]byte("wholesalers")}}}, func(key, val []byte) error {
		return json.Unmarshal(val, &tmp)
	})

	if err == ErrNotFound {
		return nil
	} else if err != nil {
		return err
	}

	return ErrAlreadyExists
}

func (w *Wholesaler) Fetch() error {
	return db.Get([]Row{{Key: []byte(w.User.Email), Buckets: [][]byte{[]byte("wholesalers")}}}, func(key, val []byte) error {
		return json.Unmarshal(val, &w)
	})
}
