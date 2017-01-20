package store

import (
	"errors"
	"net/http"
	"path/filepath"
)

var (
	DefaultPrice string
)

type Config struct {
	BlockKey      string   `env:"STORE_BLOCK_KEY" envDefault:"yellow submarine"`
	DataDir       string   `env:"STORE_DATADIR" envDefault:"/var/log/store"`
	DefaultPrice  string   `env:"STORE_DEFAULT_PRICE" envDefault:"0.00"`
	Domains       []string `env:"STORE_DOMAINS" envDefault:"127.0.0.1"`
	Email         string   `env:"STORE_EMAIL" envDefault:"yellow submarine"`
	EmailPassword string   `env:"STORE_EMAIL_PASSWORD" envDefault:"yellow submarine"`
	HashKey       string   `env:"STORE_HASH_KEY" envDefault:"we all live in a"`
	Iface         string   `env:"STORE_IFACE" envDefault:"127.0.0.1"`
	LogOutput     string   `env:"STORE_LOGOUTPUT" envDefault:"stdout"`
	Name          string   `env:"STORE_NAME" envDefault:"yellow submarine"`
	Port          int      `env:"STORE_PORT" envDefault:"8080"`
	TLS           bool     `env:"STORE_TLS" envDefault:"false"`
	TLSCerts      string   `env:"STORE_TLS_CERTS" envDefault:"$HOME/.store/certs"`
	LetsEncrypt   bool     `env:"STORE_LETS_ENCRYPT" envDefault:"false"`

	RecaptchaSiteKey   string `env:"RECAPTCHA_SITE_KEY" envDefault:"yellow submarine"`
	RecaptchaURL       string `env:"RECAPTCHA_URL" envDefault:"yellow submarine"`
	RecpatchaSecretKey string `env:"RECAPTCHA_SECRET_KEY" envDefault:"yellow submarine"`

	ShopifyAPI    string `env:"SHOPIFY_API" envDefault:"yellow submarine"`
	ShopifyDomain string `env:"SHOPIFY_DOMAIN" envDefault:"yellow submarine"`
	ShopifyJSKey  string `env:"SHOPIFY_JS_KEY" envDefault:"yellow submarine"`
}

type Row struct {
	Key     []byte
	Buckets [][]byte
	Val     []byte
}

func NewRow(opts ...func(*Row)) Row {
	r := &Row{}
	for _, o := range opts {
		o(r)
	}
	return *r
}

func Key(k string) func(*Row) {
	return func(r *Row) {
		r.Key = []byte(k)
	}
}

func Buckets(buckets ...string) func(*Row) {
	return func(r *Row) {
		for _, b := range buckets {
			r.Buckets = append(r.Buckets, []byte(b))
		}
	}
}

func Val(v []byte) func(*Row) {
	return func(r *Row) {
		r.Val = v
	}
}

type storer interface {
	Put([]Row) error
	Get([]Row, func(k, v []byte) error) error
	GetAll(Row, func(k, v []byte) error) error
	Delete([]Row) error
	DeleteAll([]byte) error
	AddBucket(Row) error
	RenameBucket(Row, Row) error
	GetBackup(w http.ResponseWriter) error
}

var (
	ErrNotFound = errors.New("not found")
	db          storer
	cfg         Config
)

func Init(c Config, opts ...func()) {
	cfg = c
	for _, opt := range opts {
		opt()
	}

	if db == nil {
		b := getBolt(filepath.Join(cfg.DataDir, "db"))
		db = &Bolt{db: b}
	}

	DefaultPrice = cfg.DefaultPrice
}

func SetDB(d storer) func() {
	return func() {
		db = d
	}
}

func GetImage(bucket, title, size string) ([]byte, error) {
	q := []Row{{Key: []byte(size), Buckets: [][]byte{[]byte("images"), []byte(bucket), []byte(title)}}}
	var img []byte
	return img, db.Get(q, func(k, v []byte) error {
		img = v
		return nil
	})
}
