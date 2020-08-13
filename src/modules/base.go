package modules

import (
	"aniapi-go/models"
	"aniapi-go/utils"
	"errors"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Module is the basic interface for a module
type Module interface {
	Start(a *models.Anime)
	GetList(title string) *goquery.Selection
	GetTarget(s *goquery.Selection) string
	GetEpisodesNumber(s *goquery.Selection) int
	GetURL(s *goquery.Selection) string
	AddToMatches(animeID int, episodes int, ratio float64, target string, url string) *models.Matching
	GetMatches(animeID int) []models.Matching
}

// ModuleScrapeURL tries to parse an URI HTML
func ModuleScrapeURL(url string) (*goquery.Document, error) {
	transport := &http.Transport{
		Proxy: utils.GetBestProxy,
	}

	client := http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Printf("ERRORE: %s", err.Error())
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("URL (%s) REQUEST ERROR: %s", url, err.Error())
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
		log.Printf("URL (%s) REQUEST ERROR: %s", url, err.Error())
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		log.Printf("URL (%s) READING ERROR: %s", url, err.Error())
		return nil, err
	}

	return doc, nil
}

// ModuleFuzzyWuzzy matches an anime model with a specific module search results
func ModuleFuzzyWuzzy(m Module, titles []string, a *models.Anime) (string, int) {
	match := ""
	episodes := 0
	best := 99
	ratio := 0.0
	var otherMatches []*models.Matching

	for _, title := range titles {
		list := m.GetList(title)

		if list != nil {
			list.Each(func(_ int, s *goquery.Selection) {
				target := m.GetTarget(s)
				source := strings.Replace(strings.ToLower(title), ":", "", -1)

				score := fuzzy.RankMatch(source, target)

				bigger := math.Max(float64(len(source)), float64(len(target)))
				r := (bigger - float64(score)) / bigger

				eps := m.GetEpisodesNumber(s)

				if score < best && score != -1 && score <= 1 {
					best = score
					match = m.GetURL(s)

					episodes = eps

					ratio = r
				}

				if (score > 1 && len(source) > 2) || (len(source) <= 2 && score > 1 && score <= 10) {
					url := m.GetURL(s)
					otherMatches = append(otherMatches, m.AddToMatches(a.ID, eps, r, target, url))
				}
			})
		}

		if match == "" {
			for _, m := range otherMatches {
				m.Save()
			}
		}
	}

	if match == "" {
		matches := m.GetMatches(a.ID)

		if matches != nil && len(matches) > 0 {
			if matches[0].Votes > 0 {
				match = "/" + strings.Join(strings.Split(matches[0].URL, "/")[3:5], "/")
				episodes = matches[0].Episodes
			}
		}
	}

	return match, episodes
}
