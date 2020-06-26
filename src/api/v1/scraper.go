package v1

import (
	"aniapi-go/engine"
	"aniapi-go/models"
	"encoding/json"
	"net/http"
)

// ScraperHandler handle all scraper controller requests
func ScraperHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		getScraper(w, r)
	default:
		w.NotImplemented()
	}
}

func getScraper(w *engine.Response, r *engine.Request) {
	scraper, err := models.GetScraper()

	if err != nil {
		w.NotFound()
		return
	}

	json, err := json.Marshal(scraper)

	if err != nil {
		w.WriteJSONError(http.StatusInternalServerError, "Error while formatting response into JSON format")
		return
	}

	w.WriteJSON(http.StatusOK, string(json))
}
