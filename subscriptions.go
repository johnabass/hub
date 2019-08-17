package hub

import "reflect"

// subscriptions keeps track of sinks associated with a particular type of event.
// this type follows copy-on-write semantics.
type subscriptions map[reflect.Type][]sink

// add makes a clone of this subscriptions instance with the given type mapped to a
// new sink.
func (s subscriptions) add(eventType reflect.Type, newSink sink) subscriptions {
	clone := make(subscriptions)
	for k, v := range s {
		clone[k] = append([]sink{}, v...)
	}

	clone[eventType] = append(clone[eventType], newSink)
	return clone
}

// remove makes a clone of this subscriptions instance with the given event type's sink
// removed.  if the tuple of eventType and oldSink do not exist in this subscriptions,
// this instance is returned without modification.
func (s subscriptions) remove(eventType reflect.Type, oldSink sink) subscriptions {
	existing, ok := s[eventType]
	if !ok {
		return s
	}

	var updated []sink
	for _, candidate := range existing {
		if candidate != oldSink {
			updated = append(updated, candidate)
		}
	}

	if len(existing) == len(updated) {
		return s
	}

	clone := make(subscriptions)
	for k, v := range s {
		if k != eventType {
			clone[k] = append([]sink{}, v...)
		} else {
			clone[k] = updated
		}
	}

	return clone
}

// publish broadcasts the given event to the appropriate sinks
func (s subscriptions) publish(e interface{}) {
	v := reflect.ValueOf(e)
	for _, sink := range s[reflect.TypeOf(e)] {
		sink.send(v)
	}
}
