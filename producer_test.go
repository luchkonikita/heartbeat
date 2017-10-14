package main

import "testing"

func TestProducerPerform(t *testing.T) {
	pipeline := make(chan URL)
	tasks := []URL{
		{"http://google.com", 0},
		{"http://yandex.ru", 0},
	}
	producer := NewProducer(pipeline)

	go producer.Perform(tasks)

	for _, url := range tasks {
		pushedUrl := <-pipeline
		if pushedUrl != url {
			t.Errorf("Expected %v to be equal %v", pushedUrl, url)
		}
	}

	_, ok := <-pipeline
	if ok {
		t.Error("Expected pipeline channel to receive 2 messages and be closed then")
	}
}
