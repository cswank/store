package config

type Config struct {
	CookieBlockKey     string   `env:"STORE_BLOCK_KEY" envDefault:"yellow submarine"`
	CookieHashKey      string   `env:"STORE_HASH_KEY" envDefault:"we all live in a"`
	DataDir            string   `env:"STORE_DATADIR" envDefault:"/var/log/store"`
	DefaultPrice       string   `env:"STORE_DEFAULT_PRICE" envDefault:"0.00"`
	Domains            []string `env:"STORE_DOMAINS" envDefault:"localhost"`
	Email              string   `env:"STORE_EMAIL" envDefault:"support@localhost"`
	EmailPassword      string   `env:"STORE_EMAIL_PASSWORD" envDefault:"hushhush"`
	Iface              string   `env:"STORE_IFACE" envDefault:"127.0.0.1"`
	LogOutput          string   `env:"STORE_LOG_OUTPUT" envDefault:"stdout"`
	Port               int      `env:"STORE_PORT" envDefault:"8080"`
	RecaptchaSecretKey string   `env:"STORE_RECAPTCHA_SECRET_KEY" envDefault:""`
	RecaptchaSiteKey   string   `env:"STORE_RECAPTCHA_SITE_KEY" envDefault:""`
	RecaptchaURL       string   `env:"STORE_RECAPTCHA_URL" envDefault:"https://www.google.com/recaptcha/api/siteverify"`
	ShopifyAPI         string   `env:"STORE_SHOPIFY_API"`
	ShopifyDomain      string   `env:"STORE_SHOPIFY_DOMAIN"`
	ShopifyJSKey       string   `env:"STORE_SHOPIFY_JS_KEY"`
	StoreName          string   `env:"STORE_NAME" envDefault:"example"`
	TLSCerts           string   `env:"STORE_TLS_CERTS"`
	UseTLS             bool     `env:"STORE_TLS" envDefault:"false"`
	UseLetsEncrypt     bool     `env:"STORE_LETS_ENCRYPT" envDefault:"true"`
	Pushes             string   `env:"STORE_PUSHES" envDefault:"{}"`
}
