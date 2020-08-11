package engine

import (
	"aniapi-go/models"
	"time"

	"github.com/gocolly/colly"
)

// QueueItem is the queue's item data definition
type QueueItem struct {
	Anime         *models.Anime `json:"anime"`
	InsertionDate time.Time     `json:"insertion_date"`
	Running       bool          `json:"running"`
	Completed     bool          `json:"completed"`
}

var scraper *Scraper = NewScraper()

// QueueItems are the queue items to elaborate progressively
var QueueItems []*QueueItem

// StartQueue starts the queue's time-related elaboration process
func StartQueue() {
	for {
		if len(QueueItems) > 0 {
			item := QueueItems[0]

			if item != nil {
				item.Running = true

				msg := &SocketMessage{
					Channel: "queue",
					Data:    item,
				}

				go SocketWriteMessage(msg)

				for _, module := range scraper.Modules {
					c := colly.NewCollector()
					SetupCollectorProxy(c)

					module.Start(item.Anime, c)
				}

				item.Completed = true

				msg = &SocketMessage{
					Channel: "queue",
					Data:    item,
				}

				go SocketWriteMessage(msg)
			}

			QueueItems = QueueItems[1:]
		}

		time.Sleep(60 * time.Second)
	}
}

// InsertItemInQueue inserts a new item at the bottom of the queue
func InsertItemInQueue(item *QueueItem) {
	QueueItems = append(QueueItems, item)

	msg := &SocketMessage{
		Channel: "queue",
		Data:    item,
	}

	go SocketWriteMessage(msg)
}

// NewQueueItem returns a new queue's item
func NewQueueItem(animeID int) *QueueItem {
	anime, err := models.GetAnime(animeID)

	if err != nil {
		return nil
	}

	return &QueueItem{
		Anime:         anime,
		InsertionDate: time.Now(),
		Running:       false,
		Completed:     false,
	}
}
