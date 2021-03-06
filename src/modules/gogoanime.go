package modules

import (
	"aniapi-go/models"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Gogoanime is the https://gogoanime.pro/ module
type Gogoanime struct{}

// GogoanimeAjax is the https://gogoanime.pro/ajax response
type GogoanimeAjax struct {
	Target string `json:"target"`
	Name   string `json:"name"`
	Error  string `json:"error"`
	HTML   string `json:"html"`
}

// Start the scraping flow
func (g Gogoanime) Start(a *models.Anime, c *colly.Collector) {
	match := g.findMatch(a, c)

	if match != "" {
		g.getEpisodes(match[len(match)-4:], a)
	}
}

func (g Gogoanime) findMatch(a *models.Anime, c *colly.Collector) string {
	titles := append([]string{a.MainTitle}, a.AlternativesTitle...)
	match := ""
	best := 99
	ratio := 0.0

	var otherMatches []*models.Matching

	for _, title := range titles {
		query := "https://gogoanime.pro/search/?language%5B%5D=subbed&keyword=" + url.QueryEscape(title)
		//c := colly.NewCollector()

		c.OnHTML("#wrapper .last_episodes .items", func(e *colly.HTMLElement) {
			e.ForEach("li", func(_ int, el *colly.HTMLElement) {
				target := strings.ToLower(el.DOM.Find(".name a").Text())
				source := strings.Replace(strings.ToLower(title), ":", "", -1)

				score := fuzzy.RankMatch(source, target)

				bigger := math.Max(float64(len(source)), float64(len(target)))
				r := (bigger - float64(score)) / bigger

				if score < best && score != -1 && score <= 1 {
					best = score
					match, _ = el.DOM.Find(".img a").Attr("href")

					ratio = r
				}

				if (score > 1 && len(source) > 2) || (len(source) <= 2 && score > 1 && score <= 10) {
					url, _ := el.DOM.Find(".showStreaming a").Eq(0).Attr("href")
					otherMatches = append(otherMatches, &models.Matching{
						AnimeID:  a.ID,
						Episodes: 0,
						From:     "gogoanime",
						Ratio:    r,
						Title:    target,
						URL:      "https://gogoanime.pro" + url,
					})
				}
			})
		})

		c.Visit(query)

		if match == "" {
			for _, m := range otherMatches {
				m.Save()
			}
		}
	}

	if match != "" {
		log.Printf("[GOGOANIME] MATCHED %s ON %s WITH %f RATIO", a.MainTitle, match, ratio)
	} else {
		matches, err := models.FindMatchings(a.ID, "gogoanime", "votes", true)

		if err == nil && len(matches) > 0 {
			if matches[0].Votes > 0 {
				match = "/" + strings.Join(strings.Split(matches[0].URL, "/")[3:5], "/")
				log.Printf("[GOGOANIME] VOTE MATCHED %s ON %s WITH %d VOTES", a.MainTitle, match, matches[0].Votes)
			}
		}
	}

	return match
}

func (g Gogoanime) getEpisodes(id string, anime *models.Anime) {
	var episodes []string
	response := &GogoanimeAjax{}

	uri := "https://gogoanime.pro/ajax/film/servers/" + id + "?ep=&episode="

	res, err := http.Get(uri)

	if err != nil {
		return
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(response)

	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(response.HTML))

	doc.Find("#episodes ul li").Each(func(_ int, s *goquery.Selection) {
		name, _ := s.Find("a").Attr("data-name")
		if name != "" {
			episodes = append(episodes, name)
		}
	})

	for _, ep := range episodes {
		ok := false

		for ok == false {
			response = &GogoanimeAjax{}
			ok = false

			number, _ := strconv.Atoi(strings.Split(ep, ":")[0])
			uri = "https://gogoanime.pro/ajax/episode/info?filmId=" + id + "&server=40&episode=" + ep + "&mcloud=9568c"

			res, err = http.Get(uri)

			if err != nil {
				continue
			}

			defer res.Body.Close()

			err = json.NewDecoder(res.Body).Decode(response)

			if err != nil {
				continue
			}

			episode := &models.Episode{
				AnimeID: anime.ID,
				From:    "gogoanime",
				Number:  number,
				Region:  models.RegionEN,
				Source:  response.Target,
				Title:   response.Name,
			}

			if episode.Source != "" {
				episode.Save()
				ok = true
			}
		}
	}
}

// NewGogoanime creates a new gogoanime module
func NewGogoanime() Gogoanime {
	return Gogoanime{}
}
