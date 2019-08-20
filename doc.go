// Package hub implements a simple publish/subscribe API for application events.  Different subsystems
// within an application can send notifications without having hard dependencies on each other.
//
// Dependency injection approaches often benefit from this approach.  Examples include sending server
// start and stop events to other components and service discovery events from libraries like consul.
//
// This package is an example of the Mediator design pattern (see https://en.wikipedia.org/wiki/Mediator_pattern).
// It is focused on managing application-layer events within a process.  It is not intended as a general
// pub/sub library for distributed notifications, although it can serve as a basis for such an implementation.
//
// Rather than requiring a specific listener interface or approach to handling events, this package supports several
// kinds of listeners via the Subscriber interface:
//
// (1) A function with exactly one input argument and no outputs.  The input argument cannot be an interface type.  The sole input type is the event type that can be passed to Publish.
//
//         h.Subscribe(func(e string) {
//             fmt.Println(e)
//         })
//
//         h.Publish("an event string")
//
// (2) A bidrectional or send-only channel.  The channel's element type cannot be an interface.  The channel's element type can be passed to Publish, which will then place it onto the channel.
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
// NOTE: If a channel subscription is cancelled, the channel passed to Subscribe is NOT closed.
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
package hub
