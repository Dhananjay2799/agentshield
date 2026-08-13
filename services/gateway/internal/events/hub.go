package events

import "sync"

type Hub struct {
	mu          sync.RWMutex
	subscribers map[chan SecurityEvent]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(
			map[chan SecurityEvent]struct{},
		),
	}
}

func (h *Hub) Subscribe() chan SecurityEvent {
	ch := make(chan SecurityEvent, 32)

	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	return ch
}

func (h *Hub) Unsubscribe(
	ch chan SecurityEvent,
) {
	h.mu.Lock()

	if _, exists := h.subscribers[ch]; exists {
		delete(h.subscribers, ch)
		close(ch)
	}

	h.mu.Unlock()
}

func (h *Hub) Broadcast(
	event SecurityEvent,
) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for subscriber := range h.subscribers {
		select {
		case subscriber <- event:
		default:
			// Do not block the Gateway if a browser
			// becomes slow or disconnected.
		}
	}
}
