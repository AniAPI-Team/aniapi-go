package v1

import "aniapi-go/engine"

// Router for api version 1
func Router(controller string, w *engine.Response, r *engine.Request) {
	switch controller {
	case "anime":
		AnimeHandler(w, r)
	case "episode":
		EpisodeHandler(w, r)
	case "matching":
		MatchingHandler(w, r)
	case "notification":
		NotificationHandler(w, r)
	case "socket":
		SocketHandler(w, r)
	default:
		w.NotFound()
	}
}
