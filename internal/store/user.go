package store

type User struct {
	Username       string
	Permission     string `json:"permission"`
	Password       string `json:"password,omitempty"`
	HashedPassword []byte `json:"hashed_password,omitempty"`
	TFA            string `json:"tfa,omitempty"`
	TFAData        []byte `json:"tfa_data,omitempty"`
	tfa            TFAer
}

func (u *User) Fetch() error {
	return nil
}
