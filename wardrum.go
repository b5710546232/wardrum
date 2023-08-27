package wardrum

import (
	"strings"
	"sync"

	"github.com/b5710546232/wardrum/internal/utils"
)

type HandleFuncType[T any] func(data T)

type Middleware[T any] func(HandleFuncType[T]) HandleFuncType[T]

type listener[T any] struct {
	HandleFunc HandleFuncType[T]
}

type wildcardListener[T any] struct {
	Pattern  string
	Listener *listener[T]
}

type Event[T any] struct {
	Name string
	Data T
}

type emitter[T any] struct {
	mu                sync.RWMutex
	listeners         map[string][]*listener[T]
	wildcardListeners []*wildcardListener[T]
	middlewares       map[string][]Middleware[T]
	historySize       int
	history           []Event[T]
}

func (emitter *emitter[T]) GetHistory() []Event[T] {
	return emitter.history
}
func (emitter *emitter[T]) applyMiddlewares(eventName string, handler HandleFuncType[T]) HandleFuncType[T] {
	if middlewares, exists := emitter.middlewares[eventName]; exists {
		for _, middleware := range middlewares {
			handler = middleware(handler)
		}
	}
	return handler

}

func (emitter *emitter[T]) getListenersFromWildcardListeners(wlds []*wildcardListener[T]) []*listener[T] {
	listeners := make([]*listener[T], len(wlds))
	for i, wld := range wlds {
		listeners[i] = wld.Listener
	}
	return listeners
}

func (emitter *emitter[T]) getMatchingWildcardListeners(eventName string) ([]*wildcardListener[T], bool) {
	matchingListeners := make([]*wildcardListener[T], 0)
	for _, wildcardListener := range emitter.wildcardListeners {
		if utils.MatchesWildcard(eventName, wildcardListener.Pattern) {
			matchingListeners = append(matchingListeners, wildcardListener)
		}
	}

	return matchingListeners, len(matchingListeners) > 0
}

type EventEmitterOptions[T any] func(*emitter[T])

func SetHistorySize[T any](size int) EventEmitterOptions[T] {
	return func(emitter *emitter[T]) {
		emitter.historySize = size
	}
}

func NewEventEmitter[T any](options ...EventEmitterOptions[T]) *emitter[T] {
	eventEmitter := &emitter[T]{
		listeners:   make(map[string][]*listener[T]),
		middlewares: make(map[string][]Middleware[T]),
		historySize: 0,
		history:     make([]Event[T], 0),
	}
	for _, opt := range options {
		opt(eventEmitter)
	}

	if eventEmitter.historySize > 0 {
		eventEmitter.history = make([]Event[T], 0, eventEmitter.historySize)
	}

	return eventEmitter
}

func NewListener[T any](handleFunc func(data T)) *listener[T] {
	return &listener[T]{HandleFunc: handleFunc}
}

func On[T any](emitter *emitter[T], eventName string, listener *listener[T]) func() {
	emitter.mu.Lock()
	defer emitter.mu.Unlock()

	if strings.Contains(eventName, "*") {
		emitter.wildcardListeners = append(emitter.wildcardListeners, &wildcardListener[T]{
			Pattern:  eventName,
			Listener: listener,
		})
		return func() {
			Off(emitter, eventName, listener)
		}
	}

	emitter.listeners[eventName] = append(emitter.listeners[eventName], listener)

	return func() {
		Off(emitter, eventName, listener)
	}
}

func Off[T any](emitter *emitter[T], eventName string, listener *listener[T]) {
	emitter.mu.Lock()
	defer emitter.mu.Unlock()
	if strings.Contains(eventName, "*") {
		for i, l := range emitter.wildcardListeners {
			if l.Listener == listener && utils.MatchesWildcard(eventName, l.Pattern) {
				emitter.wildcardListeners = append(emitter.wildcardListeners[:i], emitter.wildcardListeners[i+1:]...)
				break
			}
		}
		return
	}
	for i, l := range emitter.listeners[eventName] {
		if l == listener {
			emitter.listeners[eventName] = append(emitter.listeners[eventName][:i], emitter.listeners[eventName][i+1:]...)
			break
		}
	}
}

func Emit[T any](emitter *emitter[T], eventName string, data T) {
	emitter.mu.Lock()
	defer emitter.mu.Unlock()
	for _, listener := range emitter.listeners[eventName] {
		handlerWithMiddlewares := emitter.applyMiddlewares(eventName, listener.HandleFunc)
		handlerWithMiddlewares(data)
	}

	wildcardListeners, exists := emitter.getMatchingWildcardListeners(eventName)
	if exists {
		listeners := emitter.getListenersFromWildcardListeners(wildcardListeners)
		for _, listener := range listeners {
			handlerWithMiddlewares := emitter.applyMiddlewares(eventName, listener.HandleFunc)
			handlerWithMiddlewares(data)
		}
	}

	// Add event to history
	if len(emitter.history) >= emitter.historySize && emitter.historySize > 0 {
		emitter.history = emitter.history[1:]
	}
	emitter.history = append(emitter.history, Event[T]{Name: eventName, Data: data})
}

func Use[T any](emitter *emitter[T], eventName string, middleware Middleware[T]) {
	emitter.mu.Lock()
	defer emitter.mu.Unlock()
	emitter.middlewares[eventName] = append(emitter.middlewares[eventName], middleware)
}
