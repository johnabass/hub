package hub

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func Example() {
	h := New()

	h.Subscribe(func(r *http.Request) {
		fmt.Println(r.Method, r.RequestURI)
	})

	h.Subscribe(func(m string) {
		fmt.Println(m)
	})

	h.Publish(httptest.NewRequest("POST", "/foo", nil))
	h.Publish("an example message")

	// Output:
	// POST /foo
	// an example message
}

type ExampleEvent struct {
	Status int
}

type ExampleListener struct {
}

func (el ExampleListener) On(ee ExampleEvent) {
	fmt.Println(ee.Status)
}

func ExampleSubscribe_Struct() {
	h := New()

	h.Subscribe(ExampleListener{})

	h.Publish("nothing is listening to this")
	h.Publish(ExampleEvent{Status: 123})

	// Output:
	// 123
}
