package v1

import (
	"aniapi-go/engine"
	"aniapi-go/models"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// MatchingHandler handle all matchings controller requests
func MatchingHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		if !r.NeedSingleResource {
			getMoreMatching(w, r)
		}
	case "PUT":
		addMatching(w, r)
	case "POST":
		increaseMatchingVotes(w, r)
	default:
		w.NotImplemented()
	}
}

func getMoreMatching(w *engine.Response, r *engine.Request) {
	animeID, err := strconv.Atoi(r.Query["anime_id"])

	if err != nil {
		w.WriteJSONError(http.StatusBadRequest, "Error while converting anime id into Int32 type")
		return
	}

	from := r.Query["from"]

	sort, _ := url.QueryUnescape(r.Query["sort"])
	_, desc := r.Query["desc"]

	matchings, err := models.FindMatchings(animeID, from, sort, desc)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, err.Error())
		return
	}

	json, err := json.Marshal(matchings)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}

func addMatching(w *engine.Response, r *engine.Request) {
	matching := &models.Matching{}

	err := json.NewDecoder(r.Data.Body).Decode(matching)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting request body into JSON format")
		return
	}

	if matching.Title == "" {
		matching.Title = strings.Join(strings.Split(matching.URL, "/")[4:5], "/")
	}

	if matching.Ratio == 0 {
		matching.Ratio = 1
	}

	matching.Save()

	json, err := json.Marshal(matching)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}

func increaseMatchingVotes(w *engine.Response, r *engine.Request) {
	matching := &models.Matching{}

	err := json.NewDecoder(r.Data.Body).Decode(matching)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting request body into JSON format")
		return
	}

	err = matching.IncreaseVotes()

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while updating model")
		return
	}

	engine.InsertItemInQueue(engine.NewQueueItem(matching.AnimeID))

	w.WriteJSON(http.StatusOK, "")
}
