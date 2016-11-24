package store

import (
	"crypto"
	"fmt"

	"github.com/cswank/twofactor"
)

type TFAer interface {
	Get(string) ([]byte, []byte, error)
	Check([]byte, string) error
}

type TFA struct {
	issuer string
}

func NewTFA(issuer string) TFAer {
	return &TFA{issuer}
}

//Get generates a otp for the user, stores it in the db,
//and returns the serialized otp (needs to be saved) and PNG
//data to display for google authenticator
func (t *TFA) Get(username string) ([]byte, []byte, error) {
	otp, err := twofactor.NewTOTP(username, t.issuer, crypto.SHA1, 8)
	if err != nil {
		return nil, nil, err
	}
	data, err := otp.ToBytes()
	if err != nil {
		return nil, nil, err
	}

	qr, err := otp.QR()
	return data, qr, err
}

//Check retuns nil if the user.TFA is valid for
//that user.
func (t *TFA) Check(tfaData []byte, token string) error {
	if len(tfaData) == 0 {
		return fmt.Errorf("no 2fa data")
	}
	otp, err := twofactor.TOTPFromBytes(tfaData, t.issuer)
	if err != nil {
		return err
	}
	return otp.Validate(token)
}
