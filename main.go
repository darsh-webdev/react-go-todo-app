package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Building a Todo app with React and Go")

	err := godotenv.Load(".env")
	if err != nil{
		log.Fatal("Error loading .env file")
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal("Error connecting to MongoDB")
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB")

	collection = client.Database("golang-todo-app").Collection("todos")

	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5174"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"*"},
	}))

	router.GET("/api/todos", getTodos)
	router.POST("/api/todos", createTodo)
	router.PATCH("/api/todo/:id", updateTodo)
	router.DELETE("/api/todo/:id", deleteTodo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Fatal(router.Run("0.0.0.0:" + port))
}

func getTodos(c *gin.Context) {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		c.String(http.StatusBadRequest, "Error fetching todos")
		return
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			log.Fatal(err)
			return
		}
		todos = append(todos, todo)
	}
	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	todo := new(Todo)

	if err := c.ShouldBindBodyWithJSON(todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if todo.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todo body cannot be empty"})
		return
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating todo"})
		return
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, todo)
}

func updateTodo(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func deleteTodo(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
