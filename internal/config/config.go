package config

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
	RecaptchaSecretKey string `env:"RECAPTCHA_SECRET_KEY" envDefault:"yellow submarine"`

	ShopifyAPI    string `env:"SHOPIFY_API" envDefault:"yellow submarine"`
	ShopifyDomain string `env:"SHOPIFY_DOMAIN" envDefault:"yellow submarine"`
	ShopifyJSKey  string `env:"SHOPIFY_JS_KEY" envDefault:"yellow submarine"`
}
