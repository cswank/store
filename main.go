package main

import (
	"log"

	"github.com/GeertJohan/go.rice"
	"github.com/caarlos0/env"
	"github.com/cswank/store/internal/site"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	version = "0.0.0"
)

var (
	cfg      site.Config
	website  = kingpin.Command("site", "manage site")
	generate = website.Command("generate", "generate a site")
)

func init() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	box := rice.MustFindBox("templates")
	site.Init(cfg, box)
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version).Author("Craig Swank")
	switch kingpin.Parse() {
	case "site generate":
		err := site.Generate()
		if err != nil {
			log.Fatal(err)
		}
	}
}
