package hub

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

	h.Subscribe(func(r *http.Request) {
		fmt.Println(r.Method, r.RequestURI)
	})

	h.Subscribe(func(m string) {
		fmt.Println(m)
	})

	h.Subscribe(ExampleListener{})

	h.Publish(httptest.NewRequest("POST", "/foo", nil))
	h.Publish("an example message")
	h.Publish(ExampleEvent{Status: 123})

	// Output:
	// POST /foo
	// an example message
	// 123
}
