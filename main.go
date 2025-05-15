package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Todo struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Body      string `json:"body"`
	Completed bool   `json:"completed" gorm:"default:false"`
}

var db *gorm.DB

func main() {
	fmt.Println("Building a Todo app with React and Go")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	POSTGRESQL_URI := os.Getenv("POSTGRESQL_URI")
	dsn := POSTGRESQL_URI

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	err = db.AutoMigrate(&Todo{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Get the underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// Close the connection when you're done
	defer sqlDB.Close()

	fmt.Println("Connected to PostgreSQL")

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

	result := db.Find(&todos)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching todos: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	todo := new(Todo)

	if err := c.ShouldBindJSON(todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if todo.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todo body cannot be empty"})
		return
	}

	result := db.Create(&todo)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

func updateTodo(c *gin.Context) {
	id := c.Param("id")

	// Check is ID is valid
	var todo Todo

	result := db.First(&todo, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	// Update the completed status
	result = db.Model(&todo).Update("completed", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func deleteTodo(c *gin.Context) {
	id := c.Param("id")

	// Check is ID is valid
	var todo Todo

	result := db.First(&todo, id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	// Delete the todo
	result = db.Delete(&todo)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
