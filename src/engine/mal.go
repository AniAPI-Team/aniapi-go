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
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
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

		for s == false {
			c := colly.NewCollector(
				colly.Async(true),
			)
			SetupCollectorProxy(c)

			uri := fmt.Sprintf("https://myanimelist.net/anime.php?letter=%s&show=%d", m.letter, m.page*50)

			c.OnHTML(".js-categories-seasonal.js-block-list.list tbody", func(e *colly.HTMLElement) {
				e.ForEach("tr td .picSurround a", func(_ int, el *colly.HTMLElement) {
					start := time.Now()

					anime := m.scrapeElement(el.Attr("href"))

					elapsed := time.Since(start)

					log.Printf("Scraped %s [%d][%d] in %s", anime.MainTitle, anime.MyAnimeListID, anime.AniListID, elapsed)

					if anime.IsValid() {
						anime.Save()

						for _, module := range m.scraper.Modules {
							module.Start(anime)
						}
					}

					if anime.ID != 0 {
						m.scraper.UpdateProcess(anime)
					}

					time.Sleep(500 * time.Millisecond)
				})
			})

			c.OnError(func(_ *colly.Response, err error) {
				log.Printf(err.Error())
				s = true
			})

			log.Printf("\n\nScraping letter \"%s\" and page %d\n\n", m.letter, m.page+1)

			c.Visit(uri)

			m.page++
			time.Sleep(120 * time.Second)
		}
	}
}

func (m *MALSearch) scrapeElement(uri string) *models.Anime {
	anime := &models.Anime{}
	c := colly.NewCollector()
	SetupCollectorProxy(c)

	var wg sync.WaitGroup
	wg.Add(1)

	c.OnHTML("#contentWrapper .h1-title span", func(e *colly.HTMLElement) {
		anime.MainTitle = e.Text
	})

	c.OnHTML("#content", func(e *colly.HTMLElement) {
		col := e.DOM.Find("table tbody tr td").Eq(0).Find("div")

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
	})

	c.OnError(func(_ *colly.Response, err error) {
		return
	})

	c.OnScraped(func(_ *colly.Response) {
		wg.Done()
	})

	anime.MyAnimeListID, _ = strconv.Atoi(strings.Split(uri, "/")[4])

	c.Visit(uri)

	wg.Wait()

	m.getAnilistData(anime)

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

	resp.Body.Close()
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
