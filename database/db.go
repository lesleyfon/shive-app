package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StartDB() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error Loading the .env file")
	}

	MongoDB := os.Getenv("MONGOURI")

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		log.Fatal()
	}

	fmt.Println("Connected to MongoDB Successfully!")

	return client

}

var Client *mongo.Client = StartDB()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	MONGO_DB_NAME := os.Getenv("MONGO_DATABASE_NAME")
	var collection *mongo.Collection = client.Database(MONGO_DB_NAME).Collection(collectionName)
	return collection
}
