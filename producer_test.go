package main

import "testing"

func TestProducerPerform(t *testing.T) {
	pipeline := make(chan URL)
	tasks := []URL{
		{"http://google.com", 0},
		{"http://yandex.ru", 0},
	}
	producer := newProducer(pipeline)

	go producer.Perform(tasks)

	for _, url := range tasks {
		pushedURL := <-pipeline
		if pushedURL != url {
			t.Errorf("Expected %v to be equal %v", pushedURL, url)
		}
	}

	_, ok := <-pipeline
	if ok {
		t.Error("Expected pipeline channel to receive 2 messages and be closed then")
	}
}
