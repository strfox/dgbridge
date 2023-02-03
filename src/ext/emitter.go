package ext

import "sync"

type Event[T any] struct {
	mutex    sync.Mutex
	channels []chan T
}

// Broadcast broadcasts an item to all channels on an Event.
func (em *Event[T]) Broadcast(item T) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	for _, channel := range em.channels {
		channel <- item
	}
}

// Listen creates a new receive-only channel for an Event. The created channel
// will receive broadcast events.
func (em *Event[T]) Listen() <-chan T {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	channel := make(chan T)
	em.channels = append(em.channels, channel)
	return channel
}

// Off removes the specified channel from an Event.
func (em *Event[T]) Off(c <-chan T) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Remove all channels that are not `c`
	var filtered []chan T
	for _, s := range em.channels {
		if s != c {
			filtered = append(filtered, s)
		}
	}
	em.channels = filtered
	// `c` is no longer needed and should be garbage collected.
}
