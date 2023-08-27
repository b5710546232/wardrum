package events_test

import (
	"testing"

	"github.com/b5710546232/wardrum/events"
)

func TestOn(t *testing.T) {
	emitter := events.NewEventEmitter()

	// Define a flag to check if the handler was called
	handlerCalled := false

	// Define a test event name and a test listener
	eventName := "testEvent"
	testListener := &events.Listener{
		Handler: func(data interface{}) {
			// Set the flag to true to indicate that the handler was called
			handlerCalled = true
		},
	}

	// Call the On function to register the listener
	emitter.On(eventName, testListener)

	// Emit the event
	emitter.Emit(eventName, nil)
	// Check if the handler was called
	if !handlerCalled {
		t.Errorf("Handler was not called for event %s", eventName)
	}
}
