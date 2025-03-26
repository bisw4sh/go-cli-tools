package main

import (
	"fmt"
	"log"
	"net/http"

	"go-task/db"
	"go-task/handlers"
	"go-task/middleware"
)

func main() {
	// Initialize database
	database, err := db.InitDB("./todos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Create router and register routes
	router := http.NewServeMux()

	// Register routes with middleware
	todoHandler := handlers.NewTodoHandler(database)

	// Apply middleware to routes
	router.Handle("/todos", middleware.Chain(
		todoHandler.HandleTodos,
		middleware.Logging(),
		middleware.ContentTypeJSON(),
	))

	router.Handle("/todos/", middleware.Chain(
		todoHandler.HandleTodo,
		middleware.Logging(),
		middleware.ContentTypeJSON(),
	))

	// Start the server
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
