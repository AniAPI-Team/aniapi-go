package models

import (
	"aniapi-go/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Scraper is the MongoDB model of the scraper document
type Scraper struct {
	AnimeID   int       `bson:"anime_id" json:"anime_id"`
	StartTime time.Time `bson:"start_time" json:"start_time"`
}

// ScraperCollectionName is a string value of scraper MongoDB capped collection name
var ScraperCollectionName string = "scraper"

// Save create or update a notification model on MongoDB
func (s *Scraper) Save() {
	ctx := database.GetContext(10)
	_, _ = database.GetCollection(ScraperCollectionName).InsertOne(ctx, s)
}

// GetScraper returns the scraper model
func GetScraper() (*Scraper, error) {
	scraper := &Scraper{}

	ctx := database.GetContext(10)
	cur, err := database.GetCollection(ScraperCollectionName).Find(ctx, bson.M{})

	if err != nil {
		return scraper, err
	}

	defer cur.Close(ctx)

	cur.Next(ctx)
	err = cur.Decode(&scraper)

	if err != nil {
		return scraper, err
	}

	return scraper, nil
}
