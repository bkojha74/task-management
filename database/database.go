// database.go
// Author: Bipin Kumar Ojha (Freelancer)

package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variables to store the MongoDB client and collection references
var (
	MongoClient     *mongo.Client
	UsersCollection *mongo.Collection
	TasksCollection *mongo.Collection
)

// Init initializes the MongoDB connection and sets up the collections
// mongoURI is the URI string for connecting to the MongoDB instance
func Init(mongoURI string) {
	// Set up client options with the provided MongoDB URI
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}

	// Create a context with a timeout for the ping operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Ensure the cancel function is called to avoid context leak

	// Ping the MongoDB server to ensure the connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB: ", err)
	}

	// Assign the connected client to the global MongoClient variable
	MongoClient = client
	// Initialize the users collection reference
	UsersCollection = client.Database("taskmanager").Collection("users")
	// Initialize the tasks collection reference
	TasksCollection = client.Database("taskmanager").Collection("tasks")

	log.Println("Connected to MongoDB!")
}

// Disconnect disconnects from the MongoDB server
func Disconnect() {
	// Check if the MongoClient is not nil (i.e., it has been initialized)
	if MongoClient != nil {
		// Disconnect from MongoDB
		err := MongoClient.Disconnect(context.Background())
		if err != nil {
			log.Fatal("Error disconnecting from MongoDB: ", err)
		}
		log.Println("Disconnected from MongoDB.")
	}
}
