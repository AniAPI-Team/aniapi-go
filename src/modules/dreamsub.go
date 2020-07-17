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
	match, count := findMatch(titles, a.ID)

	if match != "" {
		if count == 1 {
			episode := &models.Episode{
				AnimeID: a.ID,
				From:    "dreamsub",
				Number:  1,
				Region:  models.RegionIT,
				Title:   "",
			}
			getSource(match, a, episode)
			episode.Save()
		} else {
			getEpisodes(match, a)
		}
	}
}

func findMatch(titles []string, id int) (string, int) {
	match := ""
	episodes := 0
	best := 99
	ratio := 0.0

	var otherMatches []*models.Matching

	for _, title := range titles {
		query := "https://dreamsub.stream/search/?q=" + url.QueryEscape(title)
		c := colly.NewCollector()

		c.OnHTML("#main-content .goblock", func(e *colly.HTMLElement) {
			e.ForEach(".tvBlock", func(_ int, el *colly.HTMLElement) {
				target := strings.ToLower(el.DOM.Find(".tvTitle .title").Text())
				source := strings.Replace(strings.ToLower(title), ":", "", -1)

				score := fuzzy.RankMatch(source, target)

				bigger := math.Max(float64(len(source)), float64(len(target)))
				r := (bigger - float64(score)) / bigger

				desc, _ := el.DOM.Find(".desc").Html()
				part := strings.Split(desc, "<br/>")[1]
				eps, _ := strconv.Atoi(strings.Replace(strings.TrimSpace(strings.Replace(strings.Split(part, ",")[0], "<b>Episodi</b>:", "", 1)), "+", "", 1))

				if score < best && score != -1 && score <= 1 {
					best = score
					match, _ = el.DOM.Find(".showStreaming a").Eq(0).Attr("href")

					episodes = eps

					ratio = r
				}

				if (score > 1 && len(source) > 2) || (len(source) <= 2 && score > 1 && score <= 10) {
					url, _ := el.DOM.Find(".showStreaming a").Eq(0).Attr("href")
					otherMatches = append(otherMatches, &models.Matching{
						AnimeID:  id,
						Episodes: eps,
						From:     "dreamsub",
						Ratio:    r,
						Title:    target,
						URL:      "https://dreamsub.stream" + url,
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
		log.Printf("MATCHED ON %s WITH %d SCORE (%f RATIO) AND %d EPISODES", match, best, ratio, episodes)
	} else {
		matches, err := models.FindMatchings(id, "dreamsub", "votes", true)

		if err == nil && len(matches) > 0 {
			if matches[0].Votes > 0 {
				match = "/" + strings.Join(strings.Split(matches[0].URL, "/")[3:5], "/")
				episodes = matches[0].Episodes
				log.Printf("VOTE MATCHED ON %s WITH %d VOTES AND %d EPISODES", match, matches[0].Votes, matches[0].Episodes)
			}
		}
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

			if title == "TBA" {
				return true
			}

			episode := &models.Episode{
				AnimeID: anime.ID,
				From:    "dreamsub",
				Number:  1,
				Region:  models.RegionIT,
				Title:   "",
			}
			getSource(link, anime, episode)
			episode.Title = title
			episode.Number = i + 1

			episode.Save()

			return true
		})
	})

	c.Visit("https://dreamsub.stream" + uri)
}

func getSource(uri string, anime *models.Anime, episode *models.Episode) {
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

	c.OnHTML("#iFrameVideoSub", func(e *colly.HTMLElement) {
		if episode.Source == "" {
			src := e.Attr("src")

			if src != "" {
				c.Visit("https://dreamsub.stream" + src)
			}
		}
	})

	c.OnHTML("#gotVVVVID", func(e *colly.HTMLElement) {
		episode.Source = e.Attr("href")
	})

	if uri == "" {
		return
	}

	c.Visit("https://dreamsub.stream" + uri)
}

// NewDreamsub creates a new dreamsub module
func NewDreamsub() Dreamsub {
	return Dreamsub{}
}
