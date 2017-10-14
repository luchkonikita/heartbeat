package main

import "flag"
import "fmt"
import "github.com/gosuri/uiprogress"
import "github.com/olekukonko/tablewriter"
import "log"
import "strconv"
import "time"
import "os"

const (
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

	sitemap, err := RequestSitemap(sitemapURL)
	if err != nil {
		log.Fatalf("Failed to download the sitemap: %v", err)
	}

	bar.Incr()

	// Create two channels for our pipeline
	tasks := make(chan URL)
	results := make(chan URL)

	workerTimeout := time.Duration(1000000 * *timeout)

	// Spawn workers
	for w := 1; w <= *concurrency; w++ {
		worker := NewWorker(workerTimeout, tasks, results)
		go worker.Perform(func(url URL) URL {
			statusCode, err := RequestPage(url.Loc)
			if err != nil {
				fmt.Println(err)
			}
			url.StatusCode = statusCode
			return url
		})
	}

	// Spawn tasks producer
	producer := NewProducer(tasks)
	go producer.Perform(sitemap.URLS[:*limit])

	// Create a consumer and join results
	consumer := NewConsumer(results)
	report := consumer.Perform(len(sitemap.URLS[:*limit]), func() {
		bar.Incr()
	})

	uiprogress.Stop()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"No", "URL", "Status"})
	for i, url := range report {
		row := []string{strconv.Itoa(i + 1), url.Loc, strconv.Itoa(url.StatusCode)}
		table.Append(row)
	}
	table.Render()
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
