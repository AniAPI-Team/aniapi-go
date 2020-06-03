package modules

import (
	"aniapi-go/models"
	"log"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Dreamsub is the https://dreamsub.stream/ module
type Dreamsub struct{}

// Start the scraping flow
func (d Dreamsub) Start(a *models.Anime) {
	titles := append([]string{a.MainTitle}, a.AlternativesTitle...)
	match, count := findMatch(titles)

	if match != "" {
		if count == 1 {
			episode := getSource(match, a)
			episode.Save()
		} else {
			getEpisodes(match, a)
		}
	}
}

func findMatch(titles []string) (string, int) {
	match := ""
	episodes := 0
	best := 99
	ratio := 0.0

	for _, title := range titles {
		query := "https://dreamsub.stream/search/?q=" + url.QueryEscape(title)
		c := colly.NewCollector()

		c.OnHTML("#main-content .goblock", func(e *colly.HTMLElement) {
			e.ForEach(".tvBlock", func(_ int, el *colly.HTMLElement) {
				target := strings.ToLower(el.DOM.Find(".tvTitle .title").Text())
				source := strings.ToLower(title)

				score := fuzzy.RankMatch(source, target)

				bigger := math.Max(float64(len(source)), float64(len(target)))
				r := (bigger - float64(score)) / bigger

				if score < best && score != -1 && score <= 1 {
					desc, _ := el.DOM.Find(".desc").Html()
					part := strings.Split(desc, "<br/>")[1]

					best = score
					match, _ = el.DOM.Find(".showStreaming a").Eq(0).Attr("href")
					episodes, _ = strconv.Atoi(strings.Replace(strings.TrimSpace(strings.Replace(part, "<b>Episodi</b>:", "", 1)), "+", "", 1))

					ratio = r
				}

				if score != -1 && score < 6 && best != 99 {
					log.Printf("POSSIBLE MATCH ON %s WITH %d SCORE (%f RATIO)", target, score, r)
				}
			})
		})

		c.Visit(query)
	}

	if match != "" {
		log.Printf("MATCHED ON %s WITH %d SCORE (%f RATIO) AND %d EPISODES", match, best, ratio, episodes)
	}

	return match, episodes
}

func getEpisodes(uri string, anime *models.Anime) {
	c := colly.NewCollector()

	c.OnHTML("#episodes-sv", func(e *colly.HTMLElement) {
		e.ForEachWithBreak(".ep-item", func(i int, el *colly.HTMLElement) bool {
			noepisodes := el.DOM.Find("center")

			if noepisodes.Length() == 1 {
				return false
			}

			title := ""
			parts := strings.Split(el.DOM.Find(".sli-name a").Text(), ": ")

			if len(parts) > 1 {
				title = strings.TrimSpace(parts[1])
			}
			link, _ := el.DOM.Find(".sli-name a").Attr("href")

			episode := getSource(link, anime)
			episode.Title = title
			episode.Number = i + 1

			episode.Save()

			return true
		})
	})

	c.Visit("https://dreamsub.stream" + uri)
}

func getSource(uri string, anime *models.Anime) *models.Episode {
	episode := models.Episode{
		AnimeID: anime.ID,
		From:    "dreamsub",
		Number:  1,
		Region:  models.RegionIT,
		Title:   "",
	}
	c := colly.NewCollector()

	c.OnHTML("#main-content.onlyDesktop .goblock-content div", func(e *colly.HTMLElement) {
		source := ""
		max := 0

		e.DOM.Find("a.dwButton").Each(func(_ int, a *goquery.Selection) {
			quality, _ := strconv.Atoi(strings.Replace(a.Text(), "p", "", 1))

			if quality > max {
				max = quality
				source, _ = a.Attr("href")
			}
		})

		episode.Source = source
	})

	if uri != "" {
		c.Visit("https://dreamsub.stream" + uri)
	}

	return &episode
}

// NewDreamsub creates a new dreamsub module
func NewDreamsub() Dreamsub {
	return Dreamsub{}
}
