package models

import (
	"aniapi-go/database"
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
	RegionEN EpisodeRegion = "en"
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
		// TODO: AGGIUNGERE NOTIFICA NUOVO EPISODIO SU SITO SPECIFICO
	}

	return true
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
