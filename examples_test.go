package hub

import (
	"fmt"
	"sync"
)

type ExampleEvent struct {
	Status int
}

type ExampleListener struct {
}

func (el ExampleListener) On(ee ExampleEvent) {
	fmt.Println(ee.Status)
}

func Example() {
	h := New()

	h.Subscribe(func(m map[string]string) {
		fmt.Println(m["foo"])
	})

	// use a wait group to demonstrate channel closure and to ensure all output happens
	w := new(sync.WaitGroup)
	w.Add(1)

	c := make(chan string, 1)
	cancel, _ := h.Subscribe(c, func() { close(c) })
	go func() {
		defer w.Done()
		for m := range c {
			fmt.Println(m)
		}
	}()

	h.Subscribe(ExampleListener{})

	h.Publish(map[string]string{"foo": "bar"})
	h.Publish("an example message")
	h.Publish(ExampleEvent{Status: 123})

	cancel()
	w.Wait()

	// Unordered output:
	// bar
	// an example message
	// 123
}
