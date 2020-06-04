package v1

import (
	"aniapi-go/engine"
	"aniapi-go/models"
	"aniapi-go/utils"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// EpisodeHandler handle all episodes controller requests
func EpisodeHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		if r.NeedSingleResource {
			getOneEpisode(w, r)
		} else {
			getMoreEpisode(w, r)
		}
	default:
		w.NotImplemented()
	}
}

func getOneEpisode(w *engine.Response, r *engine.Request) {
	animeID, err := strconv.Atoi(r.Params[0])

	if err != nil {
		w.WriteJSONError(http.StatusBadRequest, "Error while converting anime id into Int32 type")
		return
	}

	number, err := strconv.Atoi(r.Params[1])

	if err != nil {
		w.WriteJSONError(http.StatusBadRequest, "Error while converting episode number into Int32 type")
		return
	}

	region := ""
	if len(r.Params) > 2 {
		region = r.Params[2]
	}

	episode, err := models.GetEpisode(animeID, number, region)

	if err != nil {
		w.NotFound()
		return
	}

	json, err := json.Marshal(episode)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}

func getMoreEpisode(w *engine.Response, r *engine.Request) {
	pageNumber, _ := strconv.Atoi(r.Query["page"])
	page := utils.GetPageInfo(pageNumber)

	animeID, err := strconv.Atoi(r.Query["anime_id"])

	if err != nil {
		w.WriteJSONError(http.StatusBadRequest, "Error while converting anime id into Int32 type")
		return
	}

	from, _ := url.QueryUnescape(r.Query["from"])
	region, _ := url.QueryUnescape(r.Query["region"])

	sort, _ := url.QueryUnescape(r.Query["sort"])
	_, desc := r.Query["desc"]

	episodes, err := models.FindEpisodes(animeID, from, region, page, sort, desc)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, err.Error())
		return
	}

	json, err := json.Marshal(episodes)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}
