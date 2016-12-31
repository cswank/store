package site

import "github.com/GeertJohan/go.rice"

var (
	cfg Config
	box *rice.Box
)

type Config struct {
	LogOutput    string `env:"STORE_LOGOUTPUT" envDefault:"stdout"`
	DataDir      string `env:"STORE_DATADIR" envDefault:"/var/log/store"`
	DefaultPrice string `env:"STORE_DEFAULT_PRICE" envDefault:"0.00"`
}
