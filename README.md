# War Drum - Go Event Emitter

Streamline your Golang application's event handling with our lightweight and intuitive event emitter library. Easily emit and listen to events, fostering dynamic communication between components. Simplify development and create more responsive systems with the power of event-driven architecture.



## Installation

```bash
go get github.com/b5710546232/wardrum
```

## How to use

```go
func main() {
	emitter := events.NewEventEmitter()

	emitter.On("exampleEvent", &events.Listener{
		Handler: func(data interface{}) {
			fmt.Println("Emitter - Received data:", data)
		}})


	emitter.Emit("exampleEvent", "Hello World!")
}
```