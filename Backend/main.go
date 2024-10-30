package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
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

// CreateBookHandler handles the creation of a new book.
func CreateBookHandler(c *gin.Context) {
	type Book struct {
		ISBN   string `json:"isbn"`
		Title  string `json:"title"`
		Author string `json:"author"`
	}

	var newBook Book
	if err := c.BindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	book := bson.D{
		{Key: "isbn", Value: newBook.ISBN},
		{Key: "title", Value: newBook.Title},
		{Key: "author", Value: newBook.Author},
	}

	_, err := collection.InsertOne(ctx, book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert book"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Book created successfully!"})
}

// ReadBooksHandler handles fetching all books in the collection.
func ReadBooksHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch books"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading results"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// DeleteBookHandler handles deleting a book by ISBN.
func DeleteBookHandler(c *gin.Context) {
	isbn := c.Param("isbn")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "isbn", Value: isbn}}

	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete book"})
		return
	}

	if deleteResult.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No book found with that ISBN"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}

func main() {
	// Initialize the Gin router
	r := gin.Default()

	// Define the API routes
	r.POST("/books", CreateBookHandler)     // Create a new book
	r.GET("/books", ReadBooksHandler)       // Read all books
	r.DELETE("/books/:isbn", DeleteBookHandler) // Delete a book by ISBN

	// Start the server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
