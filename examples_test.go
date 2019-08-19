package hub

import (
	"fmt"
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

	h.Subscribe(func(m string) {
		fmt.Println(m)
	})

	h.Subscribe(ExampleListener{})

	h.Publish(map[string]string{"foo": "bar"})
	h.Publish("an example message")
	h.Publish(ExampleEvent{Status: 123})

	// Output:
	// bar
	// an example message
	// 123
}
