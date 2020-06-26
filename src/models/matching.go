package models

import (
	"aniapi-go/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Matching is the MongoDB model of a matching document
type Matching struct {
	AnimeID      int                `bson:"anime_id" json:"anime_id"`
	CreationDate time.Time          `bson:"creation_date" json:"-"`
	Episodes     int                `bson:"episodes" json:"episodes"`
	From         string             `bson:"from" json:"from"`
	MongoID      primitive.ObjectID `bson:"_id" json:"-"`
	Ratio        float64            `bson:"ratio" json:"ratio"`
	Title        string             `bson:"title" json:"title"`
	UpdateDate   time.Time          `bson:"update_date" json:"-"`
	URL          string             `bson:"url" json:"url"`
	Votes        int                `bson:"votes" json:"votes"`
}

// MatchingCollectionName is a string value of matchings MongoDB collection name
var MatchingCollectionName string = "matchings"

// IsValid checks if a matching model has the following props:
// - no duplicate
func (m *Matching) IsValid() bool {
	filter := bson.M{
		"anime_id": m.AnimeID,
		"from":     m.From,
		"title":    m.Title,
	}

	ref := &Matching{}
	ctx := database.GetContext(10)
	err := database.GetCollection(MatchingCollectionName).FindOne(ctx, filter).Decode(&ref)

	if err == nil {
		ref.URL = m.URL
		ref.Episodes = m.Episodes
		*m = *ref
	}

	return true
}

// Save create or update a matching model on MongoDB
func (m *Matching) Save() {
	if !m.IsValid() {
		return
	}

	if m.MongoID == primitive.NilObjectID {
		m.MongoID = primitive.NewObjectID()
		m.CreationDate = time.Now()
		m.Votes = 0

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(MatchingCollectionName).InsertOne(ctx, m)
	} else {
		m.UpdateDate = time.Now()

		filter := bson.M{
			"anime_id": m.AnimeID,
			"from":     m.From,
			"title":    m.Title,
		}

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(MatchingCollectionName).UpdateOne(ctx, filter, bson.M{"$set": m})
	}
}

// IncreaseVotes increases a matching existing model votes count
func (m *Matching) IncreaseVotes() error {
	filter := bson.M{
		"anime_id": m.AnimeID,
		"from":     m.From,
		"title":    m.Title,
	}

	ctx := database.GetContext(10)
	_, err := database.GetCollection(MatchingCollectionName).UpdateOne(ctx, filter, bson.M{
		"$inc": bson.M{
			"votes": 1,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

// FindMatchings returns a paginated list of filtered matchings
func FindMatchings(animeID int, from string, sort string, desc bool) ([]Matching, error) {
	var matchings []Matching

	filter := bson.M{
		"anime_id": animeID,
	}

	if from != "" {
		filter["from"] = bson.M{
			"$regex": primitive.Regex{
				Pattern: ".*" + from + ".*", Options: "i",
			},
		}
	}

	pagination := &options.FindOptions{}

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
	cur, err := database.GetCollection(MatchingCollectionName).Find(ctx, filter, pagination)

	if err != nil {
		return matchings, err
	}

	defer cur.Close(ctx)

	i := 0
	for cur.Next(ctx) {
		m := &Matching{}
		err = cur.Decode(m)

		if err != nil {
			return matchings, err
		}

		matchings = append(matchings, *m)

		i++
	}

	if len(matchings) == 0 {
		matchings = make([]Matching, 0)
	}

	return matchings, nil
}
