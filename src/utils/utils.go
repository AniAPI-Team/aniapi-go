package utils

import (
	"context"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gocolly/colly"
)

// PageInfo contains a specific page indexes information
type PageInfo struct {
	End    int
	Number int
	Start  int
	Size   int
}

var proxies []*url.URL
var proxiesUses []int
var pageSize int = 10

func LoadProxies() {
	host := os.Getenv("PROXY_HOST")
	port := os.Getenv("PROXY_PORT")
	user := os.Getenv("PROXY_USER")
	password := os.Getenv("PROXY_PASSWORD")
	count, _ := strconv.Atoi(os.Getenv("PROXY_COUNT"))

	if len(proxies) > 0 {
		return
	}

	for i := 1; i <= count; i++ {
		u := "http://" + user + "-" + strconv.Itoa(i) + ":" + password + "@" + host + ":" + port
		parsed, err := url.Parse(u)

		if err == nil {
			proxies = append(
				proxies,
				parsed,
			)

			proxiesUses = append(proxiesUses, 0)
		}
	}
}

func GetBestProxy(pr *http.Request) (*url.URL, error) {
	selected := &url.URL{}
	best := math.MaxInt32
	bestSelected := 0
	refresh := false

	for i := 0; i < len(proxies); i++ {
		uses := proxiesUses[i]

		if uses == math.MaxInt32 {
			refresh = true
		}

		if uses < best {
			best = uses
			bestSelected = i
			selected = proxies[i]

			ctx := context.WithValue(pr.Context(), colly.ProxyURLKey, proxies[i].String())
			*pr = *pr.WithContext(ctx)
		}
	}

	if refresh == true {
		for i := 0; i < len(proxies); i++ {
			proxiesUses[i] = 0
		}
	}

	proxiesUses[bestSelected]++
	return selected, nil
}

// GetPageInfo returns a page start and end indexes
func GetPageInfo(page int) *PageInfo {
	if page < 1 {
		page = 1
	}

	return &PageInfo{
		End:    page * pageSize,
		Number: page,
		Start:  (page * pageSize) - pageSize,
		Size:   pageSize,
	}
}
