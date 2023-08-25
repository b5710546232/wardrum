package events

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type EventHandler func(interface{})

type Middleware func(EventHandler) EventHandler

type Listener struct {
	Handler EventHandler
}

type ListenerOptions struct {
}

type WildcardListener struct {
	Pattern  string
	Listener *Listener
}

type Event struct {
	Name string
	Data interface{}
}

type EventEmitter struct {
	listeners         map[string][]*Listener
	wildcardListeners []*WildcardListener
	middlewares       map[string][]Middleware
	history           []Event
	historySize       int
	lock              sync.Mutex
	wildcardHandler   EventHandler
}

type EventEmitterOptions func(*EventEmitter)

func SetHistorySize(size int) EventEmitterOptions {
	return func(emitter *EventEmitter) {
		emitter.historySize = size
	}
}

func NewEventEmitter(options ...EventEmitterOptions) *EventEmitter {
	const defaultHistorySize = 0
	eventEmitter := &EventEmitter{
		listeners:   make(map[string][]*Listener),
		middlewares: make(map[string][]Middleware),
		history:     make([]Event, 0, defaultHistorySize),
		historySize: defaultHistorySize,
	}
	for _, opt := range options {
		opt(eventEmitter)
	}

	if eventEmitter.historySize > 0 {
		eventEmitter.history = make([]Event, 0, eventEmitter.historySize)
	}
	return eventEmitter

}

func (emitter *EventEmitter) On(eventName string, listener *Listener) {
	emitter.lock.Lock()
	defer emitter.lock.Unlock()

	if strings.Contains(eventName, "*") {
		emitter.wildcardListeners = append(emitter.wildcardListeners, &WildcardListener{
			Pattern:  eventName,
			Listener: listener,
		})
		return
	}

	if _, exists := emitter.listeners[eventName]; !exists {
		emitter.listeners[eventName] = []*Listener{}
	}

	emitter.listeners[eventName] = append(emitter.listeners[eventName], listener)

	go emitter.replayHistory(eventName, listener.Handler)
}

func (emitter *EventEmitter) replayHistory(eventName string, handler EventHandler) {
	for _, event := range emitter.history {
		if event.Name == eventName {
			go handler(event.Data)
		}
	}
}

func (emitter *EventEmitter) Off(eventName string, listener *Listener) {
	emitter.lock.Lock()
	defer emitter.lock.Unlock()

	if listeners, exists := emitter.listeners[eventName]; exists {
		for i, l := range listeners {
			if l == listener {
				emitter.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
}

func (emitter *EventEmitter) Use(eventName string, middleware Middleware) {
	emitter.lock.Lock()
	defer emitter.lock.Unlock()

	if _, exists := emitter.middlewares[eventName]; !exists {
		emitter.middlewares[eventName] = []Middleware{}
	}
	emitter.middlewares[eventName] = append(emitter.middlewares[eventName], middleware)
}

func (emitter *EventEmitter) Emit(eventName string, data interface{}) {
	emitter.lock.Lock()
	defer emitter.lock.Unlock()

	// Add event to history
	if len(emitter.history) >= emitter.historySize && emitter.historySize > 0 {
		emitter.history = emitter.history[1:]
	}
	emitter.history = append(emitter.history, Event{Name: eventName, Data: data})

	emitter.executeListeners(eventName, data)
	emitter.executeWildcardListeners(eventName, data)
}
func (emitter *EventEmitter) executeListeners(eventName string, data interface{}) {
	if listeners, exists := emitter.listeners[eventName]; exists {
		var wg sync.WaitGroup
		for _, listener := range listeners {
			wg.Add(1)
			go func(l *Listener) {
				defer wg.Done()
				handlerWithMiddlewares := emitter.applyMiddlewares(eventName, l.Handler)
				handlerWithMiddlewares(data)
			}(listener)
		}
		wg.Wait()
	}
}

func (emitter *EventEmitter) executeWildcardListeners(eventName string, data interface{}) {
	var wg sync.WaitGroup
	for _, wildcardListener := range emitter.wildcardListeners {
		if matchesWildcard(eventName, wildcardListener.Pattern) {
			wg.Add(1)
			go func(wl *WildcardListener) {
				defer wg.Done()
				handlerWithMiddlewares := emitter.applyMiddlewares(eventName, wl.Listener.Handler)
				handlerWithMiddlewares(data)
			}(wildcardListener)
		}
	}
	wg.Wait()
}
func (emitter *EventEmitter) applyMiddlewares(eventName string, handler EventHandler) EventHandler {
	if middlewares, exists := emitter.middlewares[eventName]; exists {
		for _, middleware := range middlewares {
			handler = middleware(handler)
		}
	}
	return handler
}

func (emitter *EventEmitter) GetHistory() []Event {
	emitter.lock.Lock()
	defer emitter.lock.Unlock()

	historyCopy := make([]Event, len(emitter.history))
	copy(historyCopy, emitter.history)

	return historyCopy
}

func matchesWildcard(eventName, pattern string) bool {
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	match, err := regexp.MatchString("^"+regexPattern+"$", eventName)
	if err != nil {
		// Handle regex error if needed
		fmt.Println(err)
		return false
	}
	return match
}
