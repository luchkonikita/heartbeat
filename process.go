package main

import "fmt"
import "github.com/gosuri/uiprogress"
import "github.com/olekukonko/tablewriter"
import "log"
import "io"
import "os"
import "strconv"
import "time"

// URL is a structure of <url> in <sitemap>
type URL struct {
	Loc        string `xml:"loc"`
	StatusCode int
}

// Sitemap is a structure of <sitemap>
type Sitemap struct {
	URLS []URL `xml:"url"`
}

// Process - Executes the program
func process(w io.Writer, concurrency int, limit int, timeout int, sitemapURL string) {
	writesToStdout := w == os.Stdout

	if writesToStdout {
		uiprogress.Start()
	}

	// Create two channels for our pipeline
	tasks := make(chan URL)
	results := make(chan URL)
	// Create pre-configured client
	client := newClient()
	// Define timeout for workers' pool
	workerTimeout := time.Duration(1000000 * timeout)

	sitemap, err := requestSitemap(client, sitemapURL)
	if err != nil {
		log.Fatalf("Error: Failed to download the sitemap: %v", err)
	}

	if len(sitemap.URLS) == 0 {
		log.Fatalf("Error: The sitemap is empty")
	}

	var entiesNum int
	if len(sitemap.URLS) > limit {
		entiesNum = len(sitemap.URLS[:limit])
	} else {
		entiesNum = len(sitemap.URLS)
	}

	bar := makeProgressBar(entiesNum)

	// Spawn workers
	for w := 1; w <= concurrency; w++ {
		worker := newWorker(workerTimeout, tasks, results)
		go worker.Perform(func(url URL) URL {
			statusCode, err := requestPage(client, url.Loc)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			url.StatusCode = statusCode
			return url
		})
	}

	// Spawn tasks producer
	producer := newProducer(tasks)
	go producer.Perform(sitemap.URLS[:entiesNum])

	// Create a consumer and join results
	consumer := newConsumer(results)
	report := consumer.Perform(entiesNum, func() {
		if writesToStdout {
			bar.Incr()
		}
	})

	// Stop the progressbar
	if writesToStdout {
		uiprogress.Stop()
	}

	// Write a report
	drawTable(w, report)
}

// Writes a report to a table and prints it
func drawTable(w io.Writer, report []URL) {
	table := tablewriter.NewWriter(w)
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
		return fmt.Sprintf("Loading page %d / %d", b.Current()-1, total)
	})
	return bar
}
