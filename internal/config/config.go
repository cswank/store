package config

type Config struct {
	DataDir           string   `env:"STORE_DATADIR" envDefault:"/var/log/store"`
	DefaultPrice      string   `env:"STORE_DEFAULT_PRICE" envDefault:"0.00"`
	DiscountCode      string   `env:"STORE_DISCOUNT_CODE" envDefault:""`
	Domains           []string `env:"STORE_DOMAINS" envDefault:"127.0.0.1"`
	Email             string   `env:"STORE_EMAIL" envDefault:""`
	EmailPassword     string   `env:"STORE_EMAIL_PASSWORD" envDefault:""`
	HashKey           string   `env:"STORE_HASH_KEY" envDefault:"we all live in a"`
	Iface             string   `env:"STORE_IFACE" envDefault:"127.0.0.1"`
	LetsEncrypt       bool     `env:"STORE_LETS_ENCRYPT" envDefault:"false"`
	LogOutput         string   `env:"STORE_LOGOUTPUT" envDefault:"stdout"`
	Name              string   `env:"STORE_NAME" envDefault:"store"`
	Port              int      `env:"STORE_PORT" envDefault:"8080"`
	ShoppingMenu      string   `env:"STORE_SHOPPING_MENU" envDefault:"menu.js"`
	TLS               bool     `env:"STORE_TLS" envDefault:"false"`
	TLSCerts          string   `env:"STORE_TLS_CERTS" envDefault:"$HOME/.store/certs"`
	UnderConstruction bool     `env:"STORE_UNDER_CONSTRUCTION" envDefault:"false"`
	WholesalePrice    string   `env:"STORE_WHOLESALE_PRICE" envDefault:"0.00"`
	BlockKey          string   `env:"STORE_BLOCK_KEY" envDefault:"yellow submarine"`

	RecaptchaSiteKey   string `env:"RECAPTCHA_SITE_KEY" envDefault:"yellow submarine"`
	RecaptchaURL       string `env:"RECAPTCHA_URL" envDefault:"yellow submarine"`
	RecaptchaSecretKey string `env:"RECAPTCHA_SECRET_KEY" envDefault:"yellow submarine"`

	ShopifyAPI    string `env:"SHOPIFY_API" envDefault:"yellow submarine"`
	ShopifyDomain string `env:"SHOPIFY_DOMAIN" envDefault:"yellow submarine"`
	ShopifyJSKey  string `env:"SHOPIFY_JS_KEY" envDefault:"yellow submarine"`

	InvoiceStylesheet string `env:"STORE_INVOICE_STYLESHEET" envDefault:"https://127.0.0.1:8080/css/invoice.css"`

	Head  string `env:"STORE_HEAD" envDefault:"head.html"`
	Home  string `env:"STORE_HOME" envDefault:"home.html"`
	About string `env:"STORE_ABOUT" envDefault:"about.html"`

	WebhookID          string `env:"STORE_WEBHOOK_ID" envDefault:""`
	WebhookIPWhitelist string `env:"STORE_WEBHOOK_IP_WHITELIST" envDefault:""`
	WebhookScript      string `env:"STORE_WEBHOOK_SCRIPT" envDefault:""`
}
