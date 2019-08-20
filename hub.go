package hub

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Cancel is a cancellation closure for subscriptions.  Cancels are idempotent.
//
// There is no requirement to cancel subscriptions.  Use of cancellation is optional.  It is an expected use case for a listener to remain subscribed for the life of the application.
type Cancel func()

// Publisher is a sink for events.  Publisher instances are safe for concurrent use and require
// no external concurrency control.
//
// The Publisher implementation returned by New() uses copy-on-write semantics for the set of listeners
// and a lock-free (atomic value) approach to publish.  This means that concurrent publishes can happen
// and should be expected.
//
// Publishers do not require events to be handled.  If an event has no subscribers, a Publisher will simply do nothing for that event.
type Publisher interface {
	// Publish routes an arbitrary object to subscribers.
	//
	// Publish is synchronous, so listeners should not perform long-running tasks
	// without spawning a goroutine.  If any listener panics, that panic will interrupt
	// event delivery and the panic will escape the call to Publish.
	Publish(interface{})
}

// Subscriber provides event subscriptions for listeners.  Subscriber instances are safe for concurrent use.
type Subscriber interface {
	// Subscribe registers a new listener.  If an error occurs, the returned
	// Cancel will be nil.
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
//     // this will panic, as return values aren't allowed
//     hub.Must(h.Subscribe(func(MyEvent) error {})
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
