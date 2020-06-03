package database

import (
	"aniapi-go/utils"
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Conn is the active connection with MongoDB
var Conn *mongo.Client

var db string

// Init of the MongoDB connection
// Should be called once in app lifecycle
func Init() {
	url := os.Getenv("MONGODB_URL")
	db = os.Getenv("MONGODB_DB")

	if url == "" {
		log.Fatalf("MONGODB_URL env var not found, shutting down...")
		os.Exit(1)
	}

	if db == "" {
		log.Fatalf("MONGODB_DB env var not found, shutting down...")
		os.Exit(1)
	}

	ctx := GetContext(10)
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(url))
	err := client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %s\n", err)
		os.Exit(1)
	} else {
		Conn = client
		log.Print("Connected to MongoDB")
	}
}

// GetContext returns a context in which to execute MongoDB operations
func GetContext(seconds time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), seconds*time.Second)
	return ctx
}

// GetCollection returns a MongoDB collection pointer
func GetCollection(name string) *mongo.Collection {
	return Conn.Database(db).Collection(name)
}

// PaginateQuery returns a MongoDB FindOptions pointer
func PaginateQuery(page *utils.PageInfo) *options.FindOptions {
	limit := int64(page.Size)
	start := int64(page.Start)

	return &options.FindOptions{
		Limit: &limit,
		Skip:  &start,
	}
}
