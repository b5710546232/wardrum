package main

import (
	"fmt"

	"github.com/b5710546232/wardrum"
)

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
