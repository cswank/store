package mock

import (
	"net/http"
	"net/http/httptest"

	"github.com/cswank/store/internal/config"
)

var (
	ts *httptest.Server
)

func FakeRecaptcha(c config.Config) config.Config {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Write([]byte(`{"success":true}`))
		}
	}))

	c.RecaptchaURL = ts.URL
	return c
}
