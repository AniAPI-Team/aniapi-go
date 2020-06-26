package engine

import (
	"aniapi-go/models"
	"time"
)

// QueueItem is the queue's item data definition
type QueueItem struct {
	Anime         *models.Anime `json:"anime"`
	InsertionDate time.Time     `json:"insertion_date"`
	Running       bool          `json:"running"`
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

				for _, module := range scraper.Modules {
					module.Start(item.Anime)
				}

				item.Running = false
			}

			QueueItems = QueueItems[1:]
		}

		time.Sleep(60 * time.Second)
	}
}

// InsertItemInQueue inserts a new item at the bottom of the queue
func InsertItemInQueue(item *QueueItem) {
	QueueItems = append(QueueItems, item)
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
	}
}
