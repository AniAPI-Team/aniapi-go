package engine

import (
	"aniapi-go/models"
	"aniapi-go/modules"
	"context"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// ScraperModule is the basic interface for a module
type ScraperModule interface {
	Start(a *models.Anime)
}

// Scraper is the data definition of the scraper engine
type Scraper struct {
	current int
	Modules []ScraperModule
	running bool
	start   time.Time
}

var proxies []*url.URL
var proxiesUses []int

// Start initializes scraper engine workflow
func (s *Scraper) Start() {
	s.loadProxies()

	s.running = true
	s.start = time.Now()
	s.UpdateProcess(0)

	mal := NewMALSearch(s)
	mal.Start()

	s.running = false
	s.UpdateProcess(-1)

	time.Sleep(6 * time.Hour)
	s.Start()
}

// UpdateProcess updates scraper process info on MongoDB
func (s *Scraper) UpdateProcess(id int) {
	s.current = id

	ss := &models.Scraper{
		AnimeID:   s.current,
		StartTime: s.start,
	}

	ss.Save()
}

func (s *Scraper) loadProxies() {
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

func getBestProxy(pr *http.Request) (*url.URL, error) {
	selected := &url.URL{}
	best := math.MaxInt32
	bestSelected := 0

	for i := 0; i < len(proxies); i++ {
		uses := proxiesUses[i]

		if uses < best {
			best = uses
			bestSelected = i
			selected = proxies[i]

			ctx := context.WithValue(pr.Context(), colly.ProxyURLKey, proxies[i].String())
			*pr = *pr.WithContext(ctx)
		}
	}

	proxiesUses[bestSelected]++
	return selected, nil
}

// SetupCollectorProxy setup colly collector to scrape even better with proxy
func SetupCollectorProxy(c *colly.Collector) {
	extensions.RandomUserAgent(c)
	c.SetProxyFunc(getBestProxy)
}

// NewScraper creates a new scraper engine
func NewScraper() *Scraper {
	return &Scraper{
		current: 0,
		running: false,
		Modules: []ScraperModule{
			modules.NewDreamsub(),
		},
	}
}
