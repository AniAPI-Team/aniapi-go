package models

import (
	"aniapi-go/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AnimeStatus is the enumerator type of anime's status
type AnimeStatus int

const (
	// AnimeStatusUnknown mean there are not information about
	AnimeStatusUnknown AnimeStatus = -1
	// AnimeStatusFinished mean the anime has completed airing
	AnimeStatusFinished AnimeStatus = 0
	// AnimeStatusAiring mean the anime is airing right now
	AnimeStatusAiring AnimeStatus = 1
	// AnimeStatusNotYet mean the anime is not yet released
	AnimeStatusNotYet AnimeStatus = 2
)

// Anime is the MongoDB model of an anime document
type Anime struct {
	AiringStart       time.Time          `bson:"airing_start" json:"airing_from"`
	AiringEnd         time.Time          `bson:"airing_end" json:"airing_to"`
	AlternativesTitle []string           `bson:"alternatives_title" json:"other_titles"`
	AniListID         int                `bson:"anilist_id" json:"anilist_id"`
	CreationDate      time.Time          `bson:"creation_date" json:"-"`
	Genres            []string           `bson:"genres" json:"genres"`
	ID                int                `bson:"id" json:"id"`
	MainTitle         string             `bson:"main_title" json:"title"`
	MongoID           primitive.ObjectID `bson:"_id" json:"-"`
	MyAnimeListID     int                `bson:"mal_id" json:"mal_id"`
	Picture           string             `bson:"picture" json:"picture"`
	Score             float32            `bson:"score" json:"score"`
	Status            AnimeStatus        `bson:"status" json:"status"`
	Type              string             `bson:"type" json:"type"`
	UpdateDate        time.Time          `bson:"update_date" json:"-"`
}

// AnimeCollectionName is a string value of animes MongoDB collection name
var AnimeCollectionName string = "animes"

// SetStatus converts MAL status to model one
func (a *Anime) SetStatus(s string) {
	if s == "Finished Airing" {
		a.Status = AnimeStatusFinished
	} else if s == "Currently Airing" {
		a.Status = AnimeStatusAiring
	} else if s == "Not yet aired" {
		a.Status = AnimeStatusNotYet
	} else {
		a.Status = AnimeStatusUnknown
	}
}

// IsValid checks if an anime model has the following props:
// - no Hentai genre
// - no duplicate
func (a *Anime) IsValid() bool {
	valid := true

	for _, g := range a.Genres {
		if g == "Hentai" {
			valid = false
		}
	}

	filter := bson.M{
		"main_title": a.MainTitle,
	}

	ref := &Anime{}
	ctx := database.GetContext(10)
	err := database.GetCollection(AnimeCollectionName).FindOne(ctx, filter).Decode(&ref)

	if err == nil && ref.MyAnimeListID != a.MyAnimeListID {
		valid = false
	} else if err == nil && ref.MyAnimeListID == a.MyAnimeListID {
		ref.AiringStart = a.AiringStart
		ref.AiringEnd = a.AiringEnd

		for _, title := range a.AlternativesTitle {
			if !isTitleDuplicate(ref.AlternativesTitle, title) {
				ref.AlternativesTitle = append(ref.AlternativesTitle, title)
			}
		}

		ref.AniListID = a.AniListID
		ref.Genres = a.Genres
		ref.Picture = a.Picture
		ref.Score = a.Score
		ref.Status = a.Status
		// TODO: VERIFICARE CAMBIAMENTO STATO E NEL CASO AGGIUNGERE NOTIFICA
		ref.Type = a.Type
		*a = *ref
	}

	return valid
}

func isTitleDuplicate(l []string, t string) bool {
	duplicate := false

	for _, s := range l {
		if s == t {
			duplicate = true
		}
	}

	return duplicate
}

// Save create or update an anime model on MongoDB
func (a *Anime) Save() {
	if a.MongoID == primitive.NilObjectID {
		a.MongoID = primitive.NewObjectID()
		a.CreationDate = time.Now()
		a.ID = getNextAvailableID()

		if a.ID == -1 {
			return
		}

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(AnimeCollectionName).InsertOne(ctx, a)
	} else {
		a.UpdateDate = time.Now()

		filter := bson.M{
			"main_title": a.MainTitle,
		}

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(AnimeCollectionName).UpdateOne(ctx, filter, bson.M{"$set": a})
	}
}

func getNextAvailableID() int {
	ctx := database.GetContext(10)
	count, err := database.GetCollection(AnimeCollectionName).CountDocuments(ctx, bson.M{})

	if err != nil {
		return -1
	} else if count == 0 {
		return 1
	}

	limit := int64(1)
	options := &options.FindOptions{
		Limit: &limit,
		Sort: bson.M{
			"id": -1,
		},
	}

	ctx = database.GetContext(10)
	cur, err := database.GetCollection(AnimeCollectionName).Find(ctx, bson.M{}, options)

	if err != nil {
		return -1
	}

	defer cur.Close(ctx)

	exist := cur.TryNext(ctx)

	if exist {
		temp := &Anime{}
		err = cur.Decode(temp)

		if err != nil {
			return -1
		}

		return temp.ID + 1
	}

	return -1
}
