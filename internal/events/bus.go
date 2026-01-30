package events

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// EventType represents different event types in the system
type EventType string

const (
	EventResourceDiscovered EventType = "resource.discovered"
	EventResourceOptimized  EventType = "resource.optimized"
	EventAIDecisionMade     EventType = "ai.decision.made"
	EventCostAnomaly        EventType = "cost.anomaly.detected"
	EventHealthChanged      EventType = "health.changed"
	EventOODACompleted      EventType = "ooda.completed"
)

// Errors
var (
	ErrQueueFull = errors.New("event queue is full")
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// EventHandler is a function that processes events
type EventHandler func(event Event) error

// NewEvent creates a new event with the given type and data
func NewEvent(eventType EventType, source string, data map[string]interface{}) Event {
	return Event{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
	}
}

// EventBus is a simple in-memory event bus for pub/sub
type EventBus struct {
	mu         sync.RWMutex
	handlers   map[EventType][]EventHandler
	eventQueue chan Event
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewEventBus creates a new event bus
func NewEventBus(bufferSize int) *EventBus {
	ctx, cancel := context.WithCancel(context.Background())

	bus := &EventBus{
		handlers:   make(map[EventType][]EventHandler),
		eventQueue: make(chan Event, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start event processor
	go bus.processEvents()

	return bus
}

// Subscribe registers a handler for an event type
func (b *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Publish publishes an event to all subscribers
func (b *EventBus) Publish(event Event) error {
	event.Timestamp = time.Now()

	select {
	case b.eventQueue <- event:
		return nil
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
		return ErrQueueFull
	}
}

// processEvents processes events from the queue
func (b *EventBus) processEvents() {
	for {
		select {
		case event := <-b.eventQueue:
			b.handleEvent(event)
		case <-b.ctx.Done():
			return
		}
	}
}

// handleEvent dispatches event to all registered handlers
func (b *EventBus) handleEvent(event Event) {
	b.mu.RLock()
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(event); err != nil {
				// Log error (would use proper logger in production)
				_ = err
			}
		}(handler)
	}
}

// Close stops the event bus
func (b *EventBus) Close() error {
	b.cancel()
	close(b.eventQueue)
	return nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Predefined event creators
func ResourceDiscoveredEvent(resourceID, provider, resourceType string) Event {
	return NewEvent(EventResourceDiscovered, "resource-scanner", map[string]interface{}{
		"resource_id":   resourceID,
		"provider":      provider,
		"resource_type": resourceType,
	})
}

func AIDecisionEvent(tier int, model, decision string, confidence float64) Event {
	return NewEvent(EventAIDecisionMade, fmt.Sprintf("tier-%d", tier), map[string]interface{}{
		"tier":       tier,
		"model":      model,
		"decision":   decision,
		"confidence": confidence,
	})
}

func CostAnomalyEvent(resourceID string, expectedCost, actualCost float64) Event {
	return NewEvent(EventCostAnomaly, "anomaly-detector", map[string]interface{}{
		"resource_id":   resourceID,
		"expected_cost": expectedCost,
		"actual_cost":   actualCost,
		"deviation":     actualCost - expectedCost,
	})
}
