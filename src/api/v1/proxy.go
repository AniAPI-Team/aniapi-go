package v1

import (
	"aniapi-go/engine"
	"net/http/httputil"
	"net/url"
)

// ProxyHandler handle a proxy request
func ProxyHandler(w *engine.Response, r *engine.Request) {
	switch r.Data.Method {
	case "GET":
		externalURL, _ := url.QueryUnescape(r.Query["url"])
		cleanURL, _ := url.Parse(externalURL)

		proxy := httputil.NewSingleHostReverseProxy(cleanURL)

		r.Data.Header.Set("X-Forwarded-Host", r.Data.Header.Get("Host"))
		r.Data.URL.Host = cleanURL.Host
		r.Data.Header.Set("Host", "cdn.dreamsub.stream")

		referer, _ := url.QueryUnescape(r.Query["referer"])
		r.Data.Header.Set("Referer", referer)
		r.Data.Host = cleanURL.Host

		proxy.ServeHTTP(w.Writer, r.Data)
	default:
		w.NotImplemented()
	}
}
