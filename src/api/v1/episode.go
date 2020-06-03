package v1

import "aniapi-go/engine"

// EpisodeHandler handle all episodes controller requests
func EpisodeHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		if r.NeedSingleResource {
			//getOne(w, r)
		} else {
			//getMore(w, r)
		}
	default:
		w.NotImplemented()
	}
}
