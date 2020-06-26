package v1

import (
	"aniapi-go/engine"
	"aniapi-go/models"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// NotificationHandler handle all notification controller requests
func NotificationHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		getMoreNotification(w, r)
	default:
		w.NotImplemented()
	}
}

func getMoreNotification(w *engine.Response, r *engine.Request) {
	var animeIDs []int
	var anilistIDs []int

	animeIDList := strings.Split(r.Query["anime_id"], ",")
	anilistIDList := strings.Split(r.Query["anilist_id"], ",")

	for _, animeID := range animeIDList {
		if animeID == "" {
			continue
		}

		id, err := strconv.Atoi(animeID)

		if err != nil {
			w.WriteJSONError(http.StatusBadRequest, "Error while converting anime id into Int32 type")
			return
		}

		animeIDs = append(animeIDs, id)
	}

	for _, anilistID := range anilistIDList {
		if anilistID == "" {
			continue
		}

		id, err := strconv.Atoi(anilistID)

		if err != nil {
			w.WriteJSONError(http.StatusBadRequest, "Error while converting anilist id into Int32 type")
			return
		}

		anilistIDs = append(anilistIDs, id)
	}

	notifications, err := models.FindNotifications(animeIDs, anilistIDs)

	if err != nil {
		w.NotFound()
		return
	}

	json, err := json.Marshal(notifications)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}
