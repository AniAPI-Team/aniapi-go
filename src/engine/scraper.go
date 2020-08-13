package engine

import (
	"aniapi-go/models"
	"aniapi-go/modules"
	"aniapi-go/utils"
	"errors"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Scraper is the data definition of the scraper engine
type Scraper struct {
	Modules []modules.Module
	running bool
	start   time.Time
}

// ScraperInfo is the data definition of the scraper process status
type ScraperInfo struct {
	Anime     *models.Anime `json:"anime"`
	StartTime time.Time     `json:"start_time"`
}

// Start initializes scraper engine workflow
func (s *Scraper) Start() {
	s.running = true
	s.start = time.Now()
	go printMemoryUsed(s)

	mal := NewMALSearch(s)
	mal.Start()

	s.running = false
	s.UpdateProcess(nil)

	time.Sleep(6 * time.Hour)
	s.Start()
}

// UpdateProcess updates scraper process
func (s *Scraper) UpdateProcess(anime *models.Anime) {
	ss := &ScraperInfo{
		Anime:     anime,
		StartTime: s.start,
	}

	msg := &SocketMessage{
		Channel: "scraper",
		Data:    ss,
	}

	go SocketWriteMessage(msg)
}

// ScrapeURL tries to parse an URI HTML
func (s *Scraper) ScrapeURL(url string) (*goquery.Document, error) {
	transport := &http.Transport{
		Proxy: utils.GetBestProxy,
	}

	client := http.Client{
		Transport: transport,
	}

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)

	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
	}

	if err != nil {
		log.Printf("URL (%s) REQUEST ERROR: %s", url, err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		log.Printf("URL (%s) READING ERROR: %s", url, err.Error())
		return nil, err
	}

	return doc, nil
}

func printMemoryUsed(s *Scraper) {
	for s.running == true {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		log.Printf("MEMORY USAGE OF %v MB", (m.Alloc / 1024 / 1024))
		time.Sleep(5 * time.Second)
	}
}

// NewScraper creates a new scraper engine
func NewScraper() *Scraper {
	return &Scraper{
		running: false,
		Modules: []modules.Module{
			modules.NewDreamsub(),
			//modules.NewGogoanime(),
		},
	}
}
