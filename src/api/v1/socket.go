package v1

import (
	"aniapi-go/engine"
)

// SocketHandler handle a socket connection initialization
func SocketHandler(w *engine.Response, r *engine.Request) {
	engine.OnSocketConnStart(w, r)

	msg := &engine.SocketMessage{
		Channel: "queue",
		Data:    engine.QueueItems,
	}

	go engine.SocketWriteMessage(msg)
}
