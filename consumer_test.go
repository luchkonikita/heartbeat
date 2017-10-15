package main

import "testing"

func TestConsumerPerform(t *testing.T) {
	pipeline := make(chan URL)
	results := []URL{
		{"http://google.com", 200},
		{"http://yandex.ru", 200},
	}
	consumer := newConsumer(pipeline)

	go func() {
		pipeline <- results[0]
		pipeline <- results[1]
	}()

	callbackCalled := false
	report := consumer.Perform(2, func() {
		callbackCalled = true
	})

	if !callbackCalled {
		t.Error("Expected Perform to call a provided callback")
	}

	for i, url := range results {
		if report[i] != url {
			t.Errorf("Expected %v to be equal %v", report[i], url)
		}
	}
}
