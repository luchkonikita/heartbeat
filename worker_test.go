package main

import "testing"
import "time"

func TestWorkerPerform(t *testing.T) {
	url := URL{"http://google.com", 0}
	tasks := make(chan URL)
	results := make(chan URL)
	worker := newWorker(time.Duration(0), tasks, results)

	go worker.Perform(func(url URL) URL {
		return url
	})

	go func() {
		tasks <- url
	}()

	result := <-results

	if result != url {
		t.Errorf("Expected to return %v as the result. Got %v instead", result, url)
	}
}
