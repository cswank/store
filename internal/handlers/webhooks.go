package handlers

import (
	"bytes"
	"net/http"
	"os/exec"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

func GetWebhooks(ch chan bool) HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) error {
		vars := mux.Vars(req)
		id := vars["id"]
		if id == "" {
			return store.ErrNotFound
		}

		if id != cfg.WebhookID {
			return store.ErrNotFound
		}

		var out bytes.Buffer
		cmd := exec.Command(cfg.WebhookScript)
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return err
		}

		lg.Printf("webhook script %s result: %q", cfg.WebhookScript, out.String())
		ch <- true
		return nil
	}
}
