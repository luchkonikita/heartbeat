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

	// Create two channels for our pipeline
	jobs := make(chan URL)
	results := make(chan URL)

	// Spawn workers
	for w := 1; w <= *concurrency; w++ {
		worker := &Worker{time.Duration(1000000 * *timeout), jobs, results}
		go worker.Perform(func(url URL) URL {
			url.StatusCode = requestPage(url.Loc)
			return url
		})
	}

	// Spawn tasks producer
	go producer(sitemap.URLS[:*limit], jobs)

	// Listen for results and gather them
	var report Sitemap

	// TODO: Think if the final number of tasks is needed here...
	for range sitemap.URLS[:*limit] {
		url := <-results
		report.URLS = append(report.URLS, url)
		bar.Incr()
	}
	uiprogress.Stop()
}

// Worker is created with the main settings: timeout and two channels.
// First channel to read tasks from and second channels is the one to push results in.
type Worker struct {
	timeout        time.Duration
	tasksChannel   <-chan URL
	resultsChannel chan<- URL
}

// A type of function which worker uses to process tasks.
type performFn func(URL) URL

// Function expects a processor function, which should do work and return some result.
// Function accepts a value received from tasks channel.
// The value returned from a proccessor is sent to results channel.
// Worker reads from tasks channel with a timeout until the channel is closed.
func (worker Worker) Perform(processor performFn) {
	for {
		select {
		case task, ok := <-worker.tasksChannel:
			if !ok {
				return
			}
			worker.resultsChannel <- processor(task)
		case <-time.Tick(worker.timeout):
			fmt.Println("waiting...")
		}
	}
}

// Simple function to be used for pushing an array into channel.
func producer(tasks []URL, pipeline chan<- URL) {
	for _, url := range tasks {
		pipeline <- url
	}
	close(pipeline)
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
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return resp.StatusCode
}

// Downloads and parses sitemap
func requestSitemap(url string) Sitemap {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	sitemap := Sitemap{}
	xml.Unmarshal(content, &sitemap)
	return sitemap
}
