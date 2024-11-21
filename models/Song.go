package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Song struct {
	ID          string   `json:"_id" bson:"_id"`
	Title       string   `json:"title" bson:"title"`
	Description string   `json:"description" bson:"description"`
	Year        string   `json:"year" bson:"year"`
	Songs       []string `json:"songs" bson:"songs"`
	Category    string   `json:"category" bson:"category"`
	CreatedAt   string   `json:"createdAt" bson:"createdAt"`
	UpdatedAt   string   `json:"updatedAt" bson:"updatedAt"`
}

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "test"

func main() {
	// Fiber app
	app := fiber.New()

	// MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://samicreed12:chatapp123@cluster0.bqp9f.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("test")
	collection := db.Collection("songs")

	// API route to fetch songs
	app.Get("/api/songs", func(c *fiber.Ctx) error {
		var songs []Song
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch songs",
			})
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var song Song
			if err := cursor.Decode(&song); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to parse song",
				})
			}
			songs = append(songs, song)
		}

		return c.JSON(songs)
	})

	// Start the server
	log.Fatal(app.Listen(":3000"))
}
