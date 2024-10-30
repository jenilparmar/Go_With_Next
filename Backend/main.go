package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection *mongo.Collection

// ConnectDB initializes a MongoDB client and connects to the database.
func ConnectDB() *mongo.Client {
	uri := "mongodb+srv://jenilparmar:dsfkjnksdfaa@cluster0.utm2zr0.mongodb.net/" // Replace with your MongoDB URI

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Verify the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}

// Initialize MongoDB connection and collection
func init() {
	client = ConnectDB()
	collection = client.Database("Library").Collection("AssignBook")
}

// CreateBook inserts a new book document into the collection.
func CreateBook(isbn string, title string, author string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	book := bson.D{
		{Key: "isbn", Value: isbn},
		{Key: "title", Value: title},
		{Key: "author", Value: author},
	}

	insertResult, err := collection.InsertOne(ctx, book)
	if err != nil {
		log.Fatal("Error inserting book:", err)
	}
	fmt.Println("Inserted a book with ID:", insertResult.InsertedID)
}

// ReadBooks fetches and prints all books in the collection.
func ReadBooks() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal("Error finding books:", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		log.Fatal("Error reading results:", err)
	}

	fmt.Println("Books in collection:")
	for _, result := range results {
		fmt.Println(result)
	}
}

// DeleteBook removes a book document based on its ISBN.
func DeleteBook(isbn string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "isbn", Value: isbn}}

	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal("Error deleting book:", err)
	}
	fmt.Printf("Deleted %v document(s)\n", deleteResult.DeletedCount)
}

func main() {
	// Examples of using the functions
	CreateBook("1234567890", "The Great Book", "John Doe")
	ReadBooks()
	// DeleteBook("1234567890")

	defer client.Disconnect(context.Background())
}
