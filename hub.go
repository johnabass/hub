package hub

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Cancel is a cancellation closure for subscriptions.  Cancels are idempotent.
type Cancel func()

// Publisher is a sink for events.  Publisher instances are safe for concurrent use and require
// no external concurrency control.
type Publisher interface {
	// Publish routes an arbitrary object to subscribers.  If there are no listeners subscribed
	// to the given type of event, this method does nothing.
	//
	// The Publisher implementation returned by New() uses copy-on-write semantics for the set of listeners
	// and a lock-free (atomic value) approach to publish.  This means that concurrent publishes can happen
	// and should be expected.
	Publish(interface{})
}

// Subscriber provides event subscriptions for listeners.  Subscriber instances are safe for concurrent use.
// However, in general a Subscriber implementation will lock on Subscribe.  Ordinarily, this isn't an issue, since
// a typical application will only subscribe each listener once at startup.
type Subscriber interface {
	// Subscribe registers a new listener, using reflection to determine the type of event.  The listener
	// can take one of several forms:
	//
	// (1) A function with exactly (1) input argument and no outputs.  The input argument cannot be an interface type.
	//     The sole input type is the event type that can be passed to Publish.
	//
	//         h.Subscribe(func(e string) {
	//             fmt.Println(e)
	//         })
	//
	//         h.Publish("an event string")
	//
	// (2) A bidrectional or send-only channel.  The channel's element type cannot be an interface.  The channel's
	//     element type can be passed to Publish, which will then place it onto the channel.
	//
	//         c := make(chan MyEvent, 1)
	//         h.Subscribe(c)
	//         go func() {
	//             for e := range c {
	//                 fmt.Println(e)
	//             }
	//         }()
	//
	//         h.Publish(MyEvent{Message: "foo"})
	//
	//     NOTE: If a channel subscription is cancelled, the channel passed to Subscribe is NOT closed.
	//
	// (3) Any type (not just a struct) with a single method of the same signature described in (1).
	//
	//         type MyListener struct{}
	//         func (ml MyListener) On(e MyEvent) {
	//             fmt.Println(e)
	//         }
	//
	//         l := MyListener{}
	//         h.Subscribe(l)
	//
	//         h.Publish(MyEvent{Status: 123})
	//
	// Any other type passed to Subscribe results in ErrInvalidListener.
	//
	// The returned cancellation closure can be used to deregister the given listener.  The cancel is idempotent.
	// Use of the cancellation is optional.  If desired, application code leave subscriptions active
	// for the life of the application.
	Subscribe(l interface{}) (Cancel, error)
}

// Interface provides both publish and subscribe functionality
type Interface interface {
	Publisher
	Subscriber
}

// New creates a hub for both publish and subscribe.  The returned implementation is optimized around
// publishes occurring much more often than subscribes.  The typical expected use case is that subscribes
// happen once, near application startup, and publishes happen throughout an application's lifetime.
func New() Interface {
	return new(hub)
}

// Must panics if err is not nil.  This function can be used to wrap Subscribe to panic instead of
// returning an error.
//
//     h := New()
//     hub.Must(h.Subscribe(func(MyEvent) error {}) // this will panic, as return values aren't allowed
func Must(c Cancel, err error) Cancel {
	if err != nil {
		panic(err)
	}

	return c
}

// hub is the internal synchronous Dispatcher implementation
type hub struct {
	subscribeLock sync.Mutex
	subscriptions atomic.Value
}

func (h *hub) load() subscriptions {
	v, _ := h.subscriptions.Load().(subscriptions)
	return v
}

func (h *hub) store(new subscriptions) {
	h.subscriptions.Store(new)
}

func (h *hub) Publish(e interface{}) {
	h.load().publish(e)
}

func (h *hub) Subscribe(l interface{}) (Cancel, error) {
	eventType, s, err := newSink(l)
	if err != nil {
		return nil, err
	}

	h.subscribeLock.Lock()
	h.store(h.load().add(eventType, s))
	h.subscribeLock.Unlock()

	return h.cancel(eventType, s), nil
}

// cancel creates a Cancel closure that will remove the given tuple from the subscriptions
func (h *hub) cancel(eventType reflect.Type, s sink) Cancel {
	var once sync.Once
	return func() {
		once.Do(func() {
			h.subscribeLock.Lock()
			h.store(h.load().remove(eventType, s))
			h.subscribeLock.Unlock()
		})
	}
}
