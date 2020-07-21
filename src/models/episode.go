package models

import (
	"aniapi-go/database"
	"aniapi-go/utils"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EpisodeRegion is the enumerator type of episode's region
type EpisodeRegion string

const (
	// RegionIT refer to italian region
	RegionIT EpisodeRegion = "it"
	// RegionEN refer to english region
	RegionEN EpisodeRegion = "gb"
)

// Episode is the MongoDB model of an episode document
type Episode struct {
	AnimeID      int                `bson:"anime_id" json:"-"`
	CreationDate time.Time          `bson:"creation_date" json:"-"`
	From         string             `bson:"from" json:"from"`
	MongoID      primitive.ObjectID `bson:"_id" json:"-"`
	Number       int                `bson:"number" json:"number"`
	Region       EpisodeRegion      `bson:"region" json:"region"`
	Source       string             `bson:"source" json:"source"`
	Title        string             `bson:"title" json:"title"`
	UpdateDate   time.Time          `bson:"update_date" json:"-"`
}

// EpisodeCollectionName is a string value of episodes MongoDB collection name
var EpisodeCollectionName string = "episodes"

// IsValid checks if an episode model has the following props:
// - no duplicate
func (e *Episode) IsValid() bool {
	filter := bson.M{
		"anime_id": e.AnimeID,
		"from":     e.From,
		"region":   e.Region,
		"number":   e.Number,
	}

	ref := &Episode{}
	ctx := database.GetContext(10)
	err := database.GetCollection(EpisodeCollectionName).FindOne(ctx, filter).Decode(&ref)

	if err == nil {
		ref.Source = e.Source
		ref.Title = e.Title
		*e = *ref
	} else {
		a, err := GetAnime(e.AnimeID)

		if err == nil {
			n := &Notification{
				AnimeID:   a.ID,
				AnilistID: a.AniListID,
				Message:   fmt.Sprintf("Episode <b>%d</b> released on <b>%s</b>", e.Number, e.From),
				Type:      TypeEpisodeChange,
			}
			n.Save()
		}
	}

	return true
}

// GetEpisode returns an existing episode model
func GetEpisode(animeID int, number int, region string) (*Episode, error) {
	episode := &Episode{}

	filter := bson.M{
		"anime_id": animeID,
		"number":   number,
	}

	if region != "" {
		filter["region"] = region
	}

	ctx := database.GetContext(10)
	err := database.GetCollection(EpisodeCollectionName).FindOne(ctx, filter).Decode(episode)

	if err != nil {
		return episode, err
	}

	return episode, nil
}

// FindEpisodes returns a paginated list of filtered episodes
func FindEpisodes(animeID int, number int, from string, region string, page *utils.PageInfo, sort string, desc bool) ([]Episode, error) {
	episodes := make([]Episode, page.Size)

	filter := bson.M{
		"anime_id": animeID,
	}

	if number != 0 {
		filter["number"] = number
	}

	if from != "" {
		filter["from"] = bson.M{
			"$regex": primitive.Regex{
				Pattern: ".*" + from + ".*", Options: "i",
			},
		}
	}

	if region != "" {
		filter["region"] = bson.M{
			"$regex": primitive.Regex{
				Pattern: ".*" + region + ".*", Options: "i",
			},
		}
	}

	pagination := database.PaginateQuery(page)

	if sort != "" {
		direction := 1

		if desc {
			direction = -1
		}

		pagination.SetSort(bson.M{
			sort: direction,
		})
	}

	ctx := database.GetContext(10)
	cur, err := database.GetCollection(EpisodeCollectionName).Find(ctx, filter, pagination)

	if err != nil {
		return episodes, err
	}

	defer cur.Close(ctx)

	i := 0
	for cur.Next(ctx) {
		err = cur.Decode(&episodes[i])

		if err != nil {
			return episodes, err
		}

		i++
	}

	return episodes[0:i], nil
}

// Save create or update an episode model on MongoDB
func (e *Episode) Save() {
	if !e.IsValid() {
		return
	}

	if e.MongoID == primitive.NilObjectID {
		e.MongoID = primitive.NewObjectID()
		e.CreationDate = time.Now()

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(EpisodeCollectionName).InsertOne(ctx, e)
	} else {
		e.UpdateDate = time.Now()

		filter := bson.M{
			"anime_id": e.AnimeID,
			"from":     e.From,
			"region":   e.Region,
			"number":   e.Number,
		}

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(EpisodeCollectionName).UpdateOne(ctx, filter, bson.M{"$set": e})
	}
}
