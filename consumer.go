package main

// A wrapper for consuming results from a channel.
type consumer struct {
	resultsChannel <-chan URL
}

// Constructor function for creating a consumer
func NewConsumer(resultsChannel <-chan URL) consumer {
	return consumer{resultsChannel}
}

// Simple function to be used for pushing an array into channel.
// As channel might remain open, the readLimit needs to be specified.
// For each gathered result the function executes callback, which
// can be used for tracking the progress of the task.
//
// Assumed to be used as a synchronous code.
func (c consumer) Perform(readLimit int, onResult func()) []URL {
	var report []URL
	for w := 1; w <= readLimit; w++ {
		url := <-c.resultsChannel
		report = append(report, url)
		onResult()
	}
	return report
}
