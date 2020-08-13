package engine

import (
	"aniapi-go/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// MALSearch is the data definition of the MAL search engine
type MALSearch struct {
	scraper *Scraper
	letter  string
	page    int
}

// ALQuery is the data definition of the AL query
type ALQuery struct {
	Query     string         `json:"query"`
	Variables map[string]int `json:"variables"`
}

// ALResponse is the data definition of the AL query response
type ALResponse struct {
	Data ALResponsePage `json:"data"`
}

// ALResponsePage is the nested data definition of the AL query response
type ALResponsePage struct {
	Page ALResponseData `json:"Page"`
}

// ALResponseData is the nested data definition of the AL query response
type ALResponseData struct {
	Media []ALResponseMedia `json:"media"`
}

// ALResponseMedia is the nested data definition of the AL query response
type ALResponseMedia struct {
	ID int `json:"id"`
}

// Start initializes MAL search engine workflow
func (m *MALSearch) Start() {
	letters := strings.Split(".ABCDEFGHIJKLMNOPQRSTUVWXYZ", "")

	for _, m.letter = range letters {
		s := false
		m.page = 0

		log.Printf("DOING LETTER %s AND PAGE %d", m.letter, m.page+1)

		for s == false {
			uri := fmt.Sprintf("https://myanimelist.net/anime.php?letter=%s&show=%d", m.letter, m.page*50)
			doc, err := m.scraper.ScrapeURL(uri)

			if err != nil {
				s = true
			} else {
				doc.Find(".js-categories-seasonal.js-block-list.list tbody tr td .picSurround a").Each(func(_ int, s *goquery.Selection) {
					animeURL, _ := s.Attr("href")
					anime := m.scrapeElement(animeURL)

					if anime != nil {
						if anime.IsValid() {
							anime.Save()

							if anime.ID != 0 {
								m.scraper.UpdateProcess(anime)
							}

							for _, module := range m.scraper.Modules {
								module.Start(anime)
							}
						}
					}

					anime = nil
				})

				doc = nil
			}

			m.page++
		}
	}
}

func (m *MALSearch) scrapeElement(uri string) *models.Anime {
	start := time.Now()
	anime := &models.Anime{}

	doc, err := m.scraper.ScrapeURL(uri)

	if err != nil {
		return nil
	}

	col := doc.Find("#content table tbody tr td").Eq(0).Find("div")

	anime.MainTitle = doc.Find("#contentWrapper .h1-title span").Eq(0).Text()
	anime.Picture, _ = col.Find("div:nth-child(1) a img").Attr("data-src")

	col.Find(".spaceit_pad").Has("span").Each(func(_ int, s *goquery.Selection) {
		synonims := false
		s.Contents().Each(func(_ int, t *goquery.Selection) {
			title := strings.TrimSpace(t.Text())

			if !t.Is("span") && len(title) != 0 {
				if synonims == true {
					titles := strings.Split(title, ", ")
					anime.AlternativesTitle = append(anime.AlternativesTitle, titles...)
					synonims = false
				}

				anime.AlternativesTitle = append(anime.AlternativesTitle, title)
			} else if t.Is("span") && title == "Synonyms:" {
				synonims = true
			}
		})
	})

	col.Find("div").Has("span.dark_text").Each(func(_ int, s *goquery.Selection) {
		getItemByKeyword(s, anime)
	})

	anime.MyAnimeListID, _ = strconv.Atoi(strings.Split(uri, "/")[4])

	m.getAnilistData(anime)

	elapsed := time.Since(start)
	log.Printf("SCRAPED %s (%d|%d) IN %s", anime.MainTitle, anime.MyAnimeListID, anime.AniListID, elapsed)

	return anime
}

func (m *MALSearch) getAnilistData(a *models.Anime) {
	al := &ALQuery{
		Query: `query($idMal: Int) { 
			Page(page: 1, perPage: 50) {
				media(idMal: $idMal, type: ANIME) { 
					id
				}
			}
		}`,
	}
	al.Variables = make(map[string]int)
	al.Variables["idMal"] = a.MyAnimeListID

	data, _ := json.Marshal(al)

	req, _ := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Print(err.Error())
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	result := &ALResponse{}
	json.Unmarshal(body, &result)

	for _, media := range result.Data.Page.Media {
		if a.AniListID == 0 || media.ID > a.AniListID {
			a.AniListID = media.ID
		} else {
			continue
		}
	}
}

func getItemByKeyword(s *goquery.Selection, a *models.Anime) {
	html, _ := s.Html()

	if strings.Contains(html, "Type") {
		a.Type = s.Find("a").Text()
	} else if strings.Contains(html, "Score") {
		score, _ := strconv.ParseFloat(s.Find("span.score-label").Text(), 32)
		a.Score = float32(score)
	} else if strings.Contains(html, "Status") {
		status := strings.TrimSpace(strings.Split(s.Last().Text(), ":")[1])
		a.SetStatus(status)
	} else if strings.Contains(html, "Aired") {
		s.Contents().Each(func(_ int, t *goquery.Selection) {
			airing := strings.TrimSpace(t.Text())

			if !t.Is("span") && len(airing) != 0 {
				parts := strings.Split(airing, "to")

				i := -1
				for _, part := range parts {
					i++

					part = strings.TrimSpace(part)
					if part == "?" {
						continue
					}

					const form = "Jan 2, 2006"
					date, _ := time.Parse(form, part)

					if i == 0 {
						a.AiringStart = date
					} else {
						a.AiringEnd = date
					}
				}
			}
		})
	} else if strings.Contains(html, "Genres") {
		s.Contents().Each(func(_ int, t *goquery.Selection) {
			if t.Is("a") {
				a.Genres = append(a.Genres, t.Text())
			}
		})
	}
}

// NewMALSearch creates a new MAL search engine
func NewMALSearch(s *Scraper) *MALSearch {
	return &MALSearch{
		scraper: s,
		page:    0,
	}
}
