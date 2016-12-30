package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/cswank/store/internal/store"
	"github.com/justinas/alice"
)

func getUser(req *http.Request) *store.User {
	u := req.Context().Value("user")
	if u == nil {
		return nil
	}
	return u.(*store.User)
}

type ACL func(req *http.Request) bool

func Or(acls ...ACL) ACL {
	return func(req *http.Request) bool {
		for _, f := range acls {
			if f(req) {
				return true
			}
		}
		return false
	}
}

func And(acls ...ACL) ACL {
	return func(req *http.Request) bool {
		b := false
		for _, f := range acls {
			b = b && f(req)
		}
		return b
	}
}

func Admin(req *http.Request) bool {
	user := getUser(req)
	return user != nil && user.Permission == store.Admin
}

func Read(req *http.Request) bool {
	user := getUser(req)
	return user != nil && (user.Permission == store.Admin || user.Permission == store.Read)
}

func Anyone(req *http.Request) bool {
	return true
}

func Perm(f ACL) alice.Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !f(req) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Not Authorized"))
				return
			}
			h.ServeHTTP(w, req)
		})
	}
}

func Human(req *http.Request) bool {
	//need to re-use the body further down the middleware chain
	d, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(d))

	form := getCaptchaForm(req)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(d))
	return postCaptcha(form)
}

func postCaptcha(form url.Values) bool {
	if form == nil {
		return false
	}
	resp, err := http.Post(captchaURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		lg.Println("invalid captcha post", err)
		return false
	}
	defer resp.Body.Close()

	var c captchaResp
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		lg.Println("invalid captcha json", err)
		return false
	}

	return c.Success
}

func getCaptchaForm(req *http.Request) url.Values {
	if err := req.ParseForm(); err != nil {
		lg.Println("invalid captcha schema form parse", err)
		return nil
	}

	return url.Values{
		"secret":   {captchaSecretKey},
		"response": {req.FormValue("g-recaptcha-response")},
	}
}
