package v1

import (
	"aniapi-go/engine"
	"encoding/json"
	"net/http"
)

// QueueHandler handle all queue controller requests
func QueueHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		getQueue(w, r)
	default:
		w.NotImplemented()
	}
}

func getQueue(w *engine.Response, r *engine.Request) {
	queue := engine.QueueItems

	if queue == nil {
		queue = make([]*engine.QueueItem, 0)
	}

	json, err := json.Marshal(queue)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}
