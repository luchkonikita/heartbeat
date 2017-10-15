package main

import "fmt"
import "time"

// Worker is created with the main settings: timeout and two channels.
// First channel to read tasks from and second channels is the one to push results in.
type worker struct {
	timeout        time.Duration
	tasksChannel   <-chan URL
	resultsChannel chan<- URL
}

// A type of function which worker uses to process tasks.
type performFn func(URL) URL

// newWorker - Constructor function for creating a worker
func newWorker(timeout time.Duration, tasksChannel <-chan URL, resultsChannel chan<- URL) worker {
	return worker{timeout, tasksChannel, resultsChannel}
}

// Function expects a processor function, which should do work and return some result.
// Function accepts a value received from tasks channel.
// The value returned from a proccessor is sent to results channel.
// Worker reads from tasks channel with a timeout until the channel is closed.
//
// Assumed to be used as an asynchronous code - spawned with a goroutine.
func (w worker) Perform(proc performFn) {
	for {
		select {
		case task, ok := <-w.tasksChannel:
			if !ok {
				return
			}
			w.resultsChannel <- proc(task)
			time.Sleep(w.timeout)
		case <-time.Tick(w.timeout):
			fmt.Println("waiting...")
		}
	}
}
