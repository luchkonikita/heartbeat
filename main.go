package main

import "encoding/xml"
import "flag"
import "fmt"
import "github.com/gosuri/uiprogress"
import "io/ioutil"
import "log"
import "net/http"
import "time"

const (
	buffer       = 1000
	dConcurrency = 5
	dLimit       = 1000
	dTimeout     = 300
)

// URL is a structure of <url> in <sitemap>
type URL struct {
	Loc        string `xml:"loc"`
	StatusCode int
}

// Sitemap is a structure of <sitemap>
type Sitemap struct {
	URLS []URL `xml:"url"`
}

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

	uiprogress.Start()
	bar := makeProgressBar(*limit)

	sitemap := requestSitemap(sitemapURL)
	bar.Incr()

	// Create a pool of goroutines to process URLs
	jobs := make(chan URL, buffer)
	results := make(chan URL, buffer)

	for w := 1; w <= *concurrency; w++ {
		go worker(*timeout, jobs, results)
	}

	// Put URLs in the queue for the pool and close the channel
	for _, url := range sitemap.URLS[:*limit] {
		jobs <- url
	}
	close(jobs)

	// Listen for results and gather them
	var report Sitemap
	for range sitemap.URLS[:*limit] {
		url := <-results
		report.URLS = append(report.URLS, url)
		bar.Incr()
	}
	uiprogress.Stop()
}

// A main worker function, runs concurrently
func worker(timeout int, jobs <-chan URL, results chan<- URL) {
	for url := range jobs {
		url.StatusCode = requestPage(url.Loc)
		results <- url
		time.Sleep(time.Duration(1000000 * timeout))
	}
}

// Creates and configures a progressbar
func makeProgressBar(total int) *uiprogress.Bar {
	bar := uiprogress.AddBar(total + 1)
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		return b.TimeElapsedString()
	})
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		switch b.Current() {
		case 0:
			return "Loading sitemap"
		default:
			return fmt.Sprintf("Loading page %d / %d", b.Current()-1, total)
		}
	})
	return bar
}

// Requests a page by URL
func requestPage(url string) int {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	return resp.StatusCode
}

// Downloads and parses sitemap
func requestSitemap(url string) Sitemap {
	resp, _ := http.Get(url)
	content, _ := ioutil.ReadAll(resp.Body)
	sitemap := Sitemap{}
	xml.Unmarshal(content, &sitemap)
	return sitemap
}
