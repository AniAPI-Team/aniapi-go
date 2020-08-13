package modules

import (
	"aniapi-go/models"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Dreamsub is the https://dreamsub.stream/ module
type Dreamsub struct{}

// Start the scraping flow
func (d Dreamsub) Start(a *models.Anime) {
	titles := append([]string{a.MainTitle}, a.AlternativesTitle...)
	match, count := ModuleFuzzyWuzzy(&d, titles, a)

	log.Printf("[DREAMSUB] MATCHED %s ON %s WITH %d EPISODES", a.MainTitle, match, count)

	if match != "" {
		if count == 1 {
			episode := &models.Episode{
				AnimeID: a.ID,
				From:    "dreamsub",
				Number:  1,
				Region:  models.RegionIT,
				Title:   "",
			}
			d.getSource(match, a, episode)
			episode.Save()
		} else {
			d.getEpisodes(match, a)
		}
	}
}

// GetList retrieves search results list
func (d Dreamsub) GetList(title string) *goquery.Selection {
	query := "https://dreamsub.stream/search/?q=" + url.QueryEscape(title)
	doc, err := ModuleScrapeURL(query)

	if err == nil {
		return doc.Find("#main-content .goblock .tvBlock")
	}

	return nil
}

// GetTarget retrieves search result title
func (d Dreamsub) GetTarget(s *goquery.Selection) string {
	return strings.ToLower(s.Find(".tvTitle .title").Text())
}

// GetEpisodesNumber retrieves search result episodes number
func (d Dreamsub) GetEpisodesNumber(s *goquery.Selection) int {
	desc, _ := s.Find(".desc").Html()
	part := strings.Split(desc, "<br/>")[1]
	eps, _ := strconv.Atoi(strings.Replace(strings.TrimSpace(strings.Replace(strings.Split(part, ",")[0], "<b>Episodi</b>:", "", 1)), "+", "", 1))

	return eps
}

// GetURL retrieves search result episode url
func (d Dreamsub) GetURL(s *goquery.Selection) string {
	url, _ := s.Find(".showStreaming a").Eq(0).Attr("href")
	return url
}

// AddToMatches adds a search result to possible matchings
func (d Dreamsub) AddToMatches(animeID int, episodes int, ratio float64, target string, url string) *models.Matching {
	return &models.Matching{
		AnimeID:  animeID,
		Episodes: episodes,
		From:     "dreamsub",
		Ratio:    ratio,
		Title:    target,
		URL:      "https://dreamsub.stream" + url,
	}
}

// GetMatches retrieves an anime model possible matchings
func (d Dreamsub) GetMatches(animeID int) []models.Matching {
	matchings, err := models.FindMatchings(animeID, "dreamsub", "votes", true)

	if err != nil {
		return nil
	}

	return matchings
}

func (d Dreamsub) getEpisodes(uri string, anime *models.Anime) {
	doc, err := ModuleScrapeURL("https://dreamsub.stream" + uri)

	if err != nil {
		return
	}

	doc.Find("#episodes-sv .ep-item").EachWithBreak(func(i int, s *goquery.Selection) bool {
		noepisodes := s.Find("center")

		if noepisodes.Length() == 1 {
			return false
		}

		title := ""
		parts := strings.Split(s.Find(".sli-name a").Text(), ": ")

		if len(parts) > 1 {
			title = strings.TrimSpace(parts[1])
		}
		link, _ := s.Find(".sli-name a").Attr("href")

		if title == "TBA" && link == "" {
			return true
		}

		episode := &models.Episode{
			AnimeID: anime.ID,
			From:    "dreamsub",
			Number:  1,
			Region:  models.RegionIT,
			Title:   "",
		}
		d.getSource(link, anime, episode)
		episode.Title = title
		episode.Number = i + 1

		episode.Save()

		log.Printf("[DREAMSUB] SAVED EPISODE %d OF %s", i+1, anime.MainTitle)

		return true
	})
}

func (d Dreamsub) getSource(uri string, anime *models.Anime, episode *models.Episode) {
	if uri == "" {
		return
	}

	doc, err := ModuleScrapeURL("https://dreamsub.stream" + uri)

	if err != nil {
		return
	}

	main := doc.Find("#main-content.onlyDesktop .goblock-content div")

	if main.Nodes != nil {
		source := ""
		max := 0

		main.Find("a.dwButton").Each(func(_ int, s *goquery.Selection) {
			quality, _ := strconv.Atoi(strings.Replace(s.Text(), "p", "", 1))

			if quality > max {
				max = quality
				source, _ = s.Attr("href")
			}
		})

		episode.Source = source
	}

	iframe := doc.Find("#iFrameVideoSub")

	if iframe.Nodes != nil {
		if episode.Source == "" {
			src, _ := iframe.Attr("src")

			if src != "" {
				d.getSource(src, anime, episode)
				//c.Visit("https://dreamsub.stream" + src)
			}
		}
	}

	vvvvid := doc.Find("#gotVVVVID")

	if vvvvid.Nodes != nil {
		episode.Source, _ = vvvvid.Attr("href")
	}
}

// NewDreamsub creates a new dreamsub module
func NewDreamsub() Dreamsub {
	return Dreamsub{}
}
