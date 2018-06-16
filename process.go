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

	"github.com/fatih/color"
	"github.com/gosuri/uiprogress"
	"github.com/olekukonko/tablewriter"
)

// TODO: Move headers somewhere
type headers []header

func (h *headers) String() string {
	return fmt.Sprint(*h)
}

func (h *headers) Set(value string) error {
	if len(*h) > 0 {
		return errors.New("headers flag already set")
	}

	for _, hd := range regexp.MustCompile(",s{0,}").Split(value, -1) {
		data := regexp.MustCompile(":s{0,}").Split(hd, 2)
		if len(data) != 2 {
			return errors.New("headers flag is invalid")
		}
		*h = append(*h, header{name: data[0], value: data[1]})
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
func process(w io.Writer, concurrency int, limit int, timeout int, ci bool, sitemapURL string, headers []header) bool {
	writesToStdout := w == os.Stdout && !ci

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

	var urlCount int = len(sitemap.URLS)
	if limit > 0 && urlCount > limit {
		urlCount = limit
	}

	bar := makeProgressBar(urlCount)

	if ci {
		fmt.Println("Heartbeat check starting")
		fmt.Println("Sitemap URL: ", sitemapURL)
		fmt.Println("Total URLs: ", urlCount)
		fmt.Println("------------------------------")
	}

	// Spawn workers
	for w := 1; w <= concurrency; w++ {
		worker := newWorker(workerTimeout, tasks, results)
		go worker.Perform(func(url URL) URL {
			statusCode, err := requestPage(client, url.Loc, headers)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			url.StatusCode = statusCode
			return url
		})
	}

	// Spawn tasks producer
	producer := newProducer(tasks)
	go producer.Perform(sitemap.URLS[:urlCount])

	// Create a consumer and join results
	counter := 0
	consumer := newConsumer(results)
	report := consumer.Perform(urlCount, func() {
		if writesToStdout {
			bar.Incr()
		} else if ci {
			counter++
			result := <-results
			message := fmt.Sprintf("[%d/%d] %s %d", counter, urlCount, result.Loc, result.StatusCode)

			if result.StatusCode == 200 {
				color.Green(message)
			} else {
				color.Red(message)
			}
		}
	})

	// Stop the progressbar
	if writesToStdout {
		uiprogress.Stop()
	}

	var failed []URL

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
