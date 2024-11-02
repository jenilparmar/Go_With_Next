package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection *mongo.Collection
var workersCollection *mongo.Collection

// ConnectDB initializes a MongoDB client and connects to the database.
func ConnectDB() *mongo.Client {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set in .env file")
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}

// Coordinates struct
type Coordinates struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

// Book struct
type Book struct {
	ISBN   string `json:"isbn" bson:"isbn"`
	Title  string `json:"title" bson:"title"`
	Author string `json:"author" bson:"author"`
}

// Worker struct
type Worker struct {
	ImgUrl       string `json:"imgUrl" bson:"imgUrl"`
	NameOfWorker string `json:"nameOfWorker" bson:"nameOfWorker"`
}

// WorkerType struct
type WorkerType struct {
	Name                string      `json:"name" bson:"name"`
	WorkName            string      `json:"workName" bson:"workName"`
	ImgUrl              string      `json:"imgUrl" bson:"imgUrl"`
	CoordinatesOfWorker Coordinates `json:"coordinatesOfWorker" bson:"coordinatesOfWorker"`
	CostPerHour         int         `json:"costPerHour" bson:"costPerHour"`
}

// Initialize MongoDB connection and collections
func init() {
	client = ConnectDB()
	collection = client.Database("Library").Collection("AssignBook")
	workersCollection = client.Database("Library").Collection("Workers")
}

// CreateBookHandler handles creating a new book.
func CreateBookHandler(c *gin.Context) {
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

	if _, err := collection.InsertOne(ctx, book); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert book"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Book created successfully!"})
}

// ReadBooksHandler handles fetching all books.
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

// workers retrieves all workers.
func workers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := workersCollection.Find(ctx, bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch workers"})
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

// AddWorker adds a new worker.
func AddWorker(c *gin.Context) {
	var newWorker Worker
	if err := c.BindJSON(&newWorker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	worker := bson.D{
		{Key: "imgUrl", Value: newWorker.ImgUrl},
		{Key: "nameOfWorker", Value: newWorker.NameOfWorker},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := workersCollection.InsertOne(ctx, worker); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert worker"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Worker created successfully!"})
}

// giveWorkerList finds all workers by WorkName.
func giveWorkerList(c *gin.Context) {
	workerName := c.Param("workerName")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := workersCollection.Find(ctx, bson.D{{Key: "workName", Value: workerName}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching workers"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading results"})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No workers found"})
		return
	}

	c.JSON(http.StatusOK, results)
}
// addWorkerTOList adds a new worker with detailed info.
func addWorkerTOList(c *gin.Context) {
    var workerName WorkerType
    if err := c.BindJSON(&workerName); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    worker := bson.D{
        {Key: "name", Value: workerName.Name},
        {Key: "workName", Value: workerName.WorkName},
        {Key: "imgUrl", Value: workerName.ImgUrl},
        {Key: "coordinatesOfWorker", Value: workerName.CoordinatesOfWorker},
        {Key: "costPerHour", Value: workerName.CostPerHour},
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Try to insert the worker and check for errors
    result, err := workersCollection.InsertOne(ctx, worker)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert worker"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Worker created successfully!", "workerId": result.InsertedID})
}


func main() {
	r := gin.Default()

	// Define routes
	r.POST("/books", CreateBookHandler)
	r.GET("/books", ReadBooksHandler)
	r.DELETE("/books/:isbn", DeleteBookHandler)
	r.GET("/feriyo/workers", workers)
	r.POST("/feriyo/addWorkers", AddWorker)
	r.POST("/feriyo/addWorkersToList", addWorkerTOList)
	r.GET("/feriyo/getWorkerToList/:workerName", giveWorkerList)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
