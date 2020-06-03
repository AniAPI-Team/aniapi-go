package api

import (
	v1 "aniapi-go/api/v1"
	"aniapi-go/engine"
	"net/http"
	"strings"
	"time"
)

// QuotaLimit is the requests quota ip memorization structure
type QuotaLimit struct {
	requestsCount   int
	lastRequestTime time.Time
}

var quotas map[string]QuotaLimit = make(map[string]QuotaLimit)

// Router for various api versions
func Router(w *engine.Response, r *engine.Request) {
	parts := strings.Split(r.Data.URL.String(), "/")
	parts = parts[1:len(parts)]

	if len(parts) >= 3 {
		version := parts[1]
		controller := parts[2]

		if askForSingleResource(len(parts)) {
			r.Params = parts[3:]
			r.NeedSingleResource = true
		} else {
			query := strings.Split(parts[2], "?")
			controller = query[0]
			r.NeedSingleResource = false

			if len(query) > 1 {
				r.Query = getQueryParameters(query[1])
			}
		}

		if !isRequestInQuotaLimit(w, r) {
			return
		}

		switch version {
		case "v1":
			v1.Router(controller, w, r)
		default:
			w.NotFound()
		}
		//} else if parts[1] == "auth" {
		//	AuthHandler(w, r)
	} else {
		w.NotFound()
	}
}

func askForSingleResource(length int) bool {
	return length > 3
}

func getQueryParameters(query string) map[string]string {
	kv := make(map[string]string)
	params := strings.Split(query, "&")

	for _, p := range params {
		param := strings.Split(p, "=")

		if len(param) > 1 {
			kv[param[0]] = param[1]
		}
	}

	return kv
}

func isRequestInQuotaLimit(w *engine.Response, r *engine.Request) bool {
	ip := getIP(r.Data)
	q, ok := quotas[ip]

	if !ok {
		quotas[ip] = QuotaLimit{
			requestsCount:   0,
			lastRequestTime: time.Now(),
		}
		q = quotas[ip]
	}

	q.requestsCount++

	elapsed := time.Since(q.lastRequestTime).Seconds()

	if elapsed > 10 {
		q.requestsCount = 1
		q.lastRequestTime = time.Now()
	}

	if q.requestsCount > 90 {
		w.TooManyRequests()
		return false
	}

	quotas[ip] = q

	return true
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")

	if forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr
}
