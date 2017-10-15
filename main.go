package main

import "flag"
import "log"
import "os"

const (
	dConcurrency = 5
	dLimit       = 1000
	dTimeout     = 300
)

// The program loads sitemap by the specified URL and then requests all the pages listed in this sitemap.
// Requests are run in parallel with the specified concurrency. For each URL program collects data.
func main() {
	// Parse arguments and setup variables
	concurrency := flag.Int("concurrency", dConcurrency, "concurrency")
	limit := flag.Int("limit", dLimit, "limit for URLs to be checked")
	timeout := flag.Int("timeout", dTimeout, "timeout for requests")
	flag.Parse()
	args := flag.Args()

	// Fail if sitemap URL is missing
	if len(args) == 0 {
		log.Fatal("Please specify the sitemap URL and try again!")
	}

	sitemapURL := args[0]

	process(os.Stdout, *concurrency, *limit, *timeout, sitemapURL)
}
