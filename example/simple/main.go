package main

import (
	"fmt"

	"github.com/b5710546232/wardrum/events"
)

func main() {
	emitter1 := events.NewEventEmitter()
	emitter2 := events.NewEventEmitter()

	emitter1.On("exampleEvent", &events.Listener{
		Handler: func(data interface{}) {
			fmt.Println("Emitter 1 - exampleEvent - Received data:", data)
		}})

	emitter2.On("exampleEvent", &events.Listener{
		Handler: func(data interface{}) {
			fmt.Println("Emitter 2 - exampleEvent - Received data:", data)
		}})

	emitter2.On("exampleEvent", &events.Listener{
		Handler: func(data interface{}) {
			fmt.Println("Emitter 2 - exampleEvent 2 - Received data:", data)
		}})

	emitter1.Emit("exampleEvent", "Hello from Emitter 1!")
	emitter2.Emit("exampleEvent", "Hello from Emitter 2!")
}
