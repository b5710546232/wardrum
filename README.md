# Wardrum - Go Event Emitter

Streamline your Golang application's event handling with our lightweight and intuitive event emitter library. Easily emit and listen to events, fostering dynamic communication between components. Simplify development and create more responsive systems with the power of event-driven architecture.



## Installation

```bash
go get github.com/b5710546232/wardrum
```

## How to use

```go
func main() {
	emitter := wardrum.NewEventEmitter[string](
		wardrum.SetHistorySize[string](3),
	)

	listener := wardrum.NewListener(func(data string) {
		fmt.Println("Received data:", data)
	})

	wardrum.On(emitter, "exampleEvent", listener)

	wardrum.Emit(emitter, "exampleEvent", "Hello world")
}
```
