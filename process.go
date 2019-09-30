package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/olekukonko/tablewriter"
)

// TODO: Move headers somewhere
type parameters []parameter

func (h *parameters) String() string {
	return fmt.Sprint(*h)
}

func (h *parameters) Set(value string) error {
	if len(*h) > 0 {
		return errors.New("headers flag already set")
	}

	for _, hd := range regexp.MustCompile(",s{0,}").Split(value, -1) {
		data := regexp.MustCompile(":s{0,}").Split(hd, 2)
		if len(data) != 2 {
			return errors.New("headers flag is invalid")
		}
		*h = append(*h, parameter{name: data[0], value: data[1]})
	}
	return nil
}

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
func process(w io.Writer, concurrency int, limit int, timeout int, sitemapURL string, headers []parameter, query []parameter) bool {
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

	sitemap, err := requestSitemap(client, sitemapURL, headers)
	if err != nil {
		log.Fatalf("Error: Failed to download the sitemap: %v", err)
	}

	if len(sitemap.URLS) == 0 {
		log.Fatalf("Error: The sitemap is empty")
	}

	var entiesNum int
	if len(sitemap.URLS) > limit && limit > 0 {
		entiesNum = len(sitemap.URLS[:limit])
	} else {
		entiesNum = len(sitemap.URLS)
	}

	bar := makeProgressBar(entiesNum)

	// Spawn workers
	for w := 1; w <= concurrency; w++ {
		worker := newWorker(workerTimeout, tasks, results)
		go worker.Perform(func(url URL) URL {
			statusCode, err := requestPage(client, appendQuery(url.Loc, query), headers)
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

	var failed []URL

	// // Write a report
	// drawTable(w, report)

	for _, url := range report {
		if url.StatusCode != 200 {
			failed = append(failed, url)
		}
	}

	if len(failed) > 0 {
		drawTable(w, failed)
	} else {
		fmt.Println("+-------------------+\n| NO PROBLEMS FOUND |\n+-------------------+")
	}

	return len(failed) == 0
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
	bar := uiprogress.AddBar(total)
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		return b.TimeElapsedString()
	})
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Loading page %d / %d", b.Current(), total)
	})
	return bar
}
