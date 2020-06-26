package models

import (
	"aniapi-go/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NotificationType is the enumerator type of notification's type
type NotificationType string

const (
	// TypeAnimeChange refer to anime type
	TypeAnimeChange NotificationType = "anime"
	// TypeEpisodeChange refer to episode type
	TypeEpisodeChange NotificationType = "episode"
)

// Notification is the MongoDB model of a notification document
type Notification struct {
	Anime        *Anime             `bson:"-" json:"anime"`
	AnimeID      int                `bson:"anime_id" json:"anime_id"`
	AnilistID    int                `bson:"anilist_id" json:"anilist_id"`
	CreationDate time.Time          `bson:"creation_date" json:"-"`
	Message      string             `bson:"message" json:"message"`
	MongoID      primitive.ObjectID `bson:"_id" json:"-"`
	Type         NotificationType   `bson:"type" json:"type"`
	UpdateDate   time.Time          `bson:"update_date" json:"on"`
}

// NotificationCollectionName is a string value of notifications MongoDB collection name
var NotificationCollectionName string = "notifications"

// IsValid checks if a notification model has the following props:
// - no duplicate
func (n *Notification) IsValid() bool {
	filter := bson.M{
		"anime_id":   n.AnimeID,
		"anilist_id": n.AnilistID,
		"type":       n.Type,
	}

	ref := &Notification{}
	ctx := database.GetContext(10)
	err := database.GetCollection(NotificationCollectionName).FindOne(ctx, filter).Decode(&ref)

	if err == nil {
		ref.Message = n.Message
		*n = *ref
	}

	return true
}

// Save create or update a notification model on MongoDB
func (n *Notification) Save() {
	if !n.IsValid() {
		return
	}

	if n.MongoID == primitive.NilObjectID {
		n.MongoID = primitive.NewObjectID()
		n.CreationDate = time.Now()

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(NotificationCollectionName).InsertOne(ctx, n)
	} else {
		n.UpdateDate = time.Now()

		filter := bson.M{
			"anime_id":   n.AnimeID,
			"anilist_id": n.AnilistID,
			"type":       n.Type,
		}

		ctx := database.GetContext(10)
		_, _ = database.GetCollection(NotificationCollectionName).UpdateOne(ctx, filter, bson.M{"$set": n})
	}
}

// FindNotifications returns a list of recent notifications
func FindNotifications(animeIDs []int, anilistIDs []int) ([]Notification, error) {
	var notifications []Notification

	weekago := time.Now().AddDate(0, 0, -7)

	var filterAnimeIDs bson.M
	var filterAnilistIDs bson.M

	if len(animeIDs) > 0 {
		filterAnimeIDs = bson.M{
			"$in": animeIDs,
		}
	}

	if len(anilistIDs) > 0 {
		filterAnilistIDs = bson.M{
			"$in": anilistIDs,
		}
	}

	filter := bson.M{
		"update_date": bson.M{
			"$gte": weekago,
		},
	}

	if filterAnimeIDs != nil || filterAnilistIDs != nil {
		filter["$or"] = bson.A{
			bson.M{
				"anime_id": filterAnimeIDs,
			},
			bson.M{
				"anilist_id": filterAnilistIDs,
			},
		}
	}

	pagination := &options.FindOptions{
		Sort: bson.M{
			"update_date": -1,
		},
	}

	ctx := database.GetContext(10)
	cur, err := database.GetCollection(NotificationCollectionName).Find(ctx, filter, pagination)

	if err != nil {
		return notifications, err
	}

	defer cur.Close(ctx)

	i := 0
	for cur.Next(ctx) {
		n := &Notification{}
		err = cur.Decode(n)

		if err != nil {
			return notifications, err
		}

		n.Anime, err = GetAnime(n.AnimeID)

		if err != nil {
			return notifications, err
		}

		notifications = append(notifications, *n)

		i++
	}

	if len(notifications) == 0 {
		notifications = make([]Notification, 0)
	}

	return notifications, nil
}
