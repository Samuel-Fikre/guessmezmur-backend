package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Song struct represents the song object in MongoDB
type Song struct {
	ID          string    `json:"_id" bson:"_id"`
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	Year        string    `json:"year" bson:"year"`
	Songs       []string  `json:"songs" bson:"songs"`
	Category    string    `json:"category" bson:"category"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
}

// MongoInstance struct holds MongoDB connection client and database
type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance
var mongoURI string

// Connect to MongoDB
func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to create MongoDB client:", err)
		return err
	}

	// Context for the MongoDB connection with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to the MongoDB server
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
		return err
	}

	// Store MongoDB client and database in MongoInstance
	mg = MongoInstance{
		Client: client,
		Db:     client.Database("test"),
	}

	log.Println("Successfully connected to MongoDB")
	return nil
}

func main() {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve MongoDB URI from environment variables
	mongoURI = os.Getenv("MONGO_URI") // Make sure MONGO_URI is set in .env

	// Connect to MongoDB
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	// Create a new Fiber app
	app := fiber.New()

	// Middleware for handling CORS, allow frontend URL from Vercel
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://mezmur-trivia.vercel.app", // Adjusted to allow your Vercel frontend
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// API route to fetch songs
	app.Get("/api/songs", func(c *fiber.Ctx) error {
		log.Println("Fetching songs from MongoDB...")

		var songs []Song

		// Fetch all songs from MongoDB
		cursor, err := mg.Db.Collection("songs").Find(c.Context(), bson.M{})
		if err != nil {
			log.Println("Error fetching songs:", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch songs",
			})
		}
		defer cursor.Close(c.Context())

		// Iterate over the cursor and decode each song
		for cursor.Next(c.Context()) {
			var song Song
			if err := cursor.Decode(&song); err != nil {
				log.Println("Error decoding song:", err)
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to parse song",
				})
			}
			songs = append(songs, song)
		}

		// Check for errors while iterating over the cursor
		if err := cursor.Err(); err != nil {
			log.Println("Cursor iteration error:", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Error while iterating over songs",
			})
		}

		// Log how many songs were retrieved
		log.Printf("Retrieved %d songs from MongoDB", len(songs))

		// Return the fetched songs as JSON
		return c.JSON(songs)
	})

	// Start the server on port 3000
	log.Println("Starting server on port 3000...")
	log.Fatal(app.Listen(":3000"))
}
