package config

type Config struct {
	HashKey          string   `env:"STORE_HASH_KEY" envDefault:"we all live in a"`
	BlockKey         string   `env:"STORE_BLOCK_KEY" envDefault:"yellow submarine"`
	LogOutput        string   `env:"STORE_LOG_OUTPUT" envDefault:"stdout"`
	DataDir          string   `env:"STORE_DATADIR" envDefault:"/var/log/store"`
	DefaultPrice     string   `env:"STORE_DEFAULT_PRICE" envDefault:"0.00"`
	StoreName        string   `env:"STORE_DEFAULT_NAME" envDefault:"example"`
	CaptchaURL       string   `env:"STORE_CAPTCHA_URL" envDefault:"example"`
	CaptchaSecretKey string   `env:"STORE_CAPTCHA_SECRET_KEY" envDefault:"example"`
	Iface            string   `env:"STORE_IFACE" envDefault:"127.0.0.1"`
	Port             int      `env:"STORE_PORT" envDefault:"8080"`
	Domains          []string `env:"STORE_DOMAINS" envDefault:"localhost"`
}
