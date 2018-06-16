package main

import (
	"flag"
	"log"
	"os"
)

const (
	dConcurrency = 5
	dLimit       = 0
	dTimeout     = 500
	dCi          = false
)

type header struct {
	name  string
	value string
}

// The program loads sitemap by the specified URL and then requests all the pages listed in this sitemap.
// Requests are run in parallel with the specified concurrency. For each URL program collects data.
func main() {
	// Parse arguments and setup variables
	concurrencyFlag := flag.Int("concurrency", dConcurrency, "concurrency")
	limitFlag := flag.Int("limit", dLimit, "limit for URLs to be checked")
	timeoutFlag := flag.Int("timeout", dTimeout, "timeout for requests")
	ciFlag := flag.Bool("ci", dCi, "omits progress on CI")

	var headersFlag headers
	flag.Var(&headersFlag, "headers", "headers to send together with requests")

	flag.Parse()
	args := flag.Args()

	// Fail if sitemap URL is missing
	if len(args) == 0 {
		log.Fatal("Please specify the sitemap URL and try again!")
	}

	sitemapURL := args[0]

	success := process(os.Stdout, *concurrencyFlag, *limitFlag, *timeoutFlag, *ciFlag, sitemapURL, headersFlag)

	if success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
