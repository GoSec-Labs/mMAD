package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/google/uuid"
)

// EventBus is the central event system
type EventBus struct {
	handlers   map[EventType][]EventHandler
	mu         sync.RWMutex
	bufferSize int
	eventChan  chan Event
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewEventBus creates a new event bus
func NewEventBus(bufferSize int) *EventBus {
	ctx, cancel := context.WithCancel(context.Background())

	bus := &EventBus{
		handlers:   make(map[EventType][]EventHandler),
		bufferSize: bufferSize,
		eventChan:  make(chan Event, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start event processing
	bus.wg.Add(1)
	go bus.processEvents()

	return bus
}

// Publish publishes an event synchronously
func (eb *EventBus) Publish(event Event) error {
	// Set ID and timestamp if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	// Execute handlers synchronously
	for _, handler := range handlers {
		if handler.CanHandle(event.Type) {
			if err := handler.Handle(event); err != nil {
				logger.Error("Event handler failed",
					"event_type", event.Type,
					"event_id", event.ID,
					"error", err)
				return err
			}
		}
	}

	return nil
}

// PublishAsync publishes an event asynchronously
func (eb *EventBus) PublishAsync(event Event) {
	// Set ID and timestamp if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case eb.eventChan <- event:
		// Event queued successfully
	default:
		logger.Warn("Event buffer full, dropping event",
			"event_type", event.Type,
			"event_id", event.ID)
	}
}

// Subscribe adds an event handler
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	logger.Info("Event handler subscribed",
		"event_type", eventType,
		"handler", fmt.Sprintf("%T", handler))

	return nil
}

// Unsubscribe removes an event handler
func (eb *EventBus) Unsubscribe(eventType EventType, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			logger.Info("Event handler unsubscribed",
				"event_type", eventType,
				"handler", fmt.Sprintf("%T", handler))
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// Close shuts down the event bus
func (eb *EventBus) Close() error {
	eb.cancel()
	close(eb.eventChan)
	eb.wg.Wait()

	logger.Info("Event bus closed")
	return nil
}

// processEvents processes events asynchronously
func (eb *EventBus) processEvents() {
	defer eb.wg.Done()

	for {
		select {
		case event, ok := <-eb.eventChan:
			if !ok {
				return // Channel closed
			}

			eb.mu.RLock()
			handlers := eb.handlers[event.Type]
			eb.mu.RUnlock()

			// Execute handlers
			for _, handler := range handlers {
				if handler.CanHandle(event.Type) {
					if err := handler.Handle(event); err != nil {
						logger.Error("Async event handler failed",
							"event_type", event.Type,
							"event_id", event.ID,
							"error", err)
					}
				}
			}

		case <-eb.ctx.Done():
			return
		}
	}
}

// GetStats returns event bus statistics
func (eb *EventBus) GetStats() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlerCount := 0
	for _, handlers := range eb.handlers {
		handlerCount += len(handlers)
	}

	return map[string]interface{}{
		"total_handlers": handlerCount,
		"event_types":    len(eb.handlers),
		"buffer_size":    eb.bufferSize,
		"queued_events":  len(eb.eventChan),
	}
}
