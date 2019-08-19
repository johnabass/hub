// Package hub implements a simple publish/subscribe API for application events.  Different subsystems
// within an application can send notifications without having hard dependencies on each other.
//
// Dependency injection approaches often benefit from this approach.  Examples include sending server
// start and stop events to other components and service discovery events from libraries like consul.
//
// This package is an example of the Mediator design pattern (see https://en.wikipedia.org/wiki/Mediator_pattern).
// It is focused on managing application-layer events within a process.  It is not intended as a general
// pub/sub library for distributed notifications, although it can serve as a basis for such an implementation.
package hub
