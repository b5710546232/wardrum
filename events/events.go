package events

import (
	"fmt"
	"regexp"
	"runtime"
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
	workerSize        int
}

type EventEmitterOptions func(*EventEmitter)

func SetHistorySize(size int) EventEmitterOptions {
	return func(emitter *EventEmitter) {
		emitter.historySize = size
	}
}

func NewEventEmitter(options ...EventEmitterOptions) *EventEmitter {
	const defaultHistorySize = 0
	workerSize := runtime.NumCPU()
	eventEmitter := &EventEmitter{
		listeners:   make(map[string][]*Listener),
		middlewares: make(map[string][]Middleware),
		history:     make([]Event, 0, defaultHistorySize),
		historySize: defaultHistorySize,
		workerSize:  workerSize,
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

	emitter.replayHistory(eventName, listener.Handler)
}

func (emitter *EventEmitter) replayHistory(eventName string, handler EventHandler) {
	for _, event := range emitter.history {
		if event.Name == eventName {
			handler(event.Data)
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

	listeners, exists := emitter.listeners[eventName]
	if exists {
		emitter.executeListeners(eventName, data, listeners)
	}

	wildcardListeners, exists := emitter.getMatchingWildcardListeners(eventName)
	if exists {
		listeners := emitter.getListenersFromWildcardListeners(wildcardListeners)
		emitter.executeListeners(eventName, data, listeners)
	}
}
func (emitter *EventEmitter) executeListeners(eventName string, data interface{}, listeners []*Listener) {
	var wg sync.WaitGroup
	workerSize := emitter.workerSize
	jobQueue := make(chan *Listener, len(listeners))

	// Start worker goroutines
	for i := 0; i < workerSize; i++ {
		go func() {
			for listener := range jobQueue {
				handlerWithMiddlewares := emitter.applyMiddlewares(eventName, listener.Handler)
				handlerWithMiddlewares(data)
				wg.Done()
			}
		}()
	}

	// Add listeners to the job queue
	for _, listener := range listeners {
		wg.Add(1)
		jobQueue <- listener
	}

	close(jobQueue)
	wg.Wait()

}

func (emitter *EventEmitter) getListenersFromWildcardListeners(wlds []*WildcardListener) []*Listener {
	listeners := make([]*Listener, len(wlds))
	for i, wld := range wlds {
		listeners[i] = wld.Listener
	}
	return listeners
}

func (emitter *EventEmitter) getMatchingWildcardListeners(eventName string) ([]*WildcardListener, bool) {
	matchingListeners := make([]*WildcardListener, 0)
	for _, wildcardListener := range emitter.wildcardListeners {
		if matchesWildcard(eventName, wildcardListener.Pattern) {
			matchingListeners = append(matchingListeners, wildcardListener)
		}
	}

	return matchingListeners, len(matchingListeners) > 0
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
