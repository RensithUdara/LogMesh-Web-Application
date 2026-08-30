package service

import (
	"sync"

	"logmesh/internal/model"
)

type EventHub struct {
	mu          sync.RWMutex
	subscribers map[chan model.LogEvent]struct{}
}

func NewEventHub() *EventHub {
	return &EventHub{
		subscribers: make(map[chan model.LogEvent]struct{}),
	}
}

func (h *EventHub) Publish(event model.LogEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for subscriber := range h.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (h *EventHub) Subscribe() chan model.LogEvent {
	ch := make(chan model.LogEvent, 16)

	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	return ch
}

func (h *EventHub) Unsubscribe(ch chan model.LogEvent) {
	h.mu.Lock()
	delete(h.subscribers, ch)
	close(ch)
	h.mu.Unlock()
}
