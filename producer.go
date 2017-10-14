package main

// A wrapper for producing a tasks list from some collection.
type producer struct {
	tasksChannel chan<- URL
}

// Constructor function for creating a producer
func NewProducer(tasksChannel chan<- URL) producer {
	return producer{tasksChannel}
}

// Simple function to be used for pushing an array into channel.
//
// Assumed to be used as an asynchronous code - spawned with a goroutine.
func (p producer) Perform(tasks []URL) {
	for _, task := range tasks {
		p.tasksChannel <- task
	}
	close(p.tasksChannel)
}
