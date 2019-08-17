package hub

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Cancel is a cancellation closure for subscriptions.  Cancels are idempotent.
type Cancel func()

// Publisher is a sink for events
type Publisher interface {
	// Publish routes an arbitrary object to subscribers
	Publish(interface{})
}

// Subscriber provides event subscriptions for listeners
type Subscriber interface {
	// Subscribe registers a new listener, using reflection to determine the type of event.
	//
	// If l is a function, it must only have (1) input argument and no outputs or an error is raised.
	// The type of event is the type of the only input argument.  A publish with an object of that type
	// will be passed to the given function.
	//
	// If l is a channel, it must not be receive-only.  The type of event is the element type of the channel.
	// A publish with an object of that type will be placed onto the channel.
	//
	// The return cancellation closure can be used to deregister the given listener.
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
