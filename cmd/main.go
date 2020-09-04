package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/juju/errors"

	formation "xsky.com/sds-formation"
	"xsky.com/sds-formation/config"
)

var (
	templateFile string
	version      bool
)

func init() {
	flag.BoolVar(&version, "version", false, "Show version")
	flag.BoolVar(&config.DryRun, "dry-run", false,
		"Report resource created successfully, but not really create them")
	flag.StringVar(&config.CachePath, "cache-path", "formation_cache", "Specify cache record path")
	flag.BoolVar(&config.NoContinue, "no-continue", false, "Do not continue from last run")
	flag.StringVar(&config.Token, "t", "",
		"Specify initial token, auth token or access token for creating resource")
	flag.StringVar(&templateFile, "f", "", "The formation template file")
}

func main() {
	flag.Parse()
	if version {
		fmt.Println(formation.DetailedVersion())
		return
	}

	log.Println(formation.Version())
	if templateFile == "" {
		log.Fatal("template file is required")
	}

	stack := new(formation.Stack)
	err := stack.Init(templateFile)
	if err != nil {
		log.Fatalf("failed to init stack using template %s: %s", templateFile, errors.ErrorStack(err))
	}
	stack.Create()

	return
}
