package v1

import (
	"aniapi-go/engine"
	"aniapi-go/models"
	"aniapi-go/utils"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// AnimeHandler handle all animes controller requests
func AnimeHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		if r.NeedSingleResource {
			getOne(w, r)
		} else {
			getMore(w, r)
		}
	default:
		w.NotImplemented()
	}
}

func getOne(w *engine.Response, r *engine.Request) {
	id, err := strconv.Atoi(r.Params[0])

	if err != nil {
		w.WriteJSONError(http.StatusBadRequest, "Error while converting anime id into Int32 type")
		return
	}

	anime, err := models.GetAnime(id)

	if err != nil {
		w.NotFound()
		return
	}

	json, err := json.Marshal(anime)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}

func getMore(w *engine.Response, r *engine.Request) {
	pageNumber, _ := strconv.Atoi(r.Query["page"])
	page := utils.GetPageInfo(pageNumber)

	title, _ := url.QueryUnescape(r.Query["title"])
	g, _ := url.QueryUnescape(r.Query["genres"])
	genres := strings.Split(g, ",")

	if genres[0] == "" {
		genres = make([]string, 0)
	}

	sort, _ := url.QueryUnescape(r.Query["sort"])
	_, desc := r.Query["desc"]

	showType, _ := url.QueryUnescape(r.Query["type"])

	animes, err := models.FindAnimes(title, genres, showType, page, sort, desc)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, err.Error())
		return
	}

	json, err := json.Marshal(animes)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}
