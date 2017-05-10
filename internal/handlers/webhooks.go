package handlers

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/cswank/store/internal/store"
	"github.com/gorilla/mux"
)

func GetWebhooks(ch chan os.Signal) HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) error {
		vars := mux.Vars(req)
		if cfg.WebhookID != vars["id"] {
			return store.ErrNotFound
		}

		whitelisted, err := checkIPWhitelist(req.RemoteAddr)
		if err != nil {
			lg.Println("could not check ip whitelist", err)
			return err
		}

		if !whitelisted {
			return store.ErrNotFound
		}

		var out bytes.Buffer
		cmd := exec.Command(cfg.WebhookScript)
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return err
		}

		lg.Printf("webhook ran: %q", out.String())
		lg.Println("webhook shutting down")
		ch <- os.Interrupt
		return nil
	}
}

// CheckIPWhitelist makes sure the provided remote address (of the form IP:port) falls within the provided IP range
// (in CIDR form or a single IP address).
func checkIPWhitelist(remoteAddr string) (bool, error) {
	// Extract IP address from remote address.
	ip := remoteAddr

	if strings.LastIndex(remoteAddr, ":") != -1 {
		ip = remoteAddr[0:strings.LastIndex(remoteAddr, ":")]
	}

	ip = strings.TrimSpace(ip)

	// IPv6 addresses will likely be surrounded by [], so don't forget to remove those.
	if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
		ip = ip[1 : len(ip)-1]
	}

	parsedIP := net.ParseIP(strings.TrimSpace(ip))

	if parsedIP == nil {
		return false, fmt.Errorf("invalid IP address found in remote address '%s'", remoteAddr)
	}

	// Extract IP range in CIDR form.  If a single IP address is provided, turn it into CIDR form.
	ipRange := cfg.WebhookIPWhitelist
	if strings.Index(ipRange, "/") == -1 {
		ipRange = ipRange + "/32"
	}

	_, cidr, err := net.ParseCIDR(ipRange)

	if err != nil {
		return false, err
	}

	return cidr.Contains(parsedIP), nil
}
