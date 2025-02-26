package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Todo represents a task in our application
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

func main() {
	// Initialize database
	initDB()
	defer db.Close()

	// Set up HTTP routes
	http.HandleFunc("/todos", handleTodos)
	http.HandleFunc("/todos/", handleTodo)

	// Start the server
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	var err error

	// Check if the database file exists
	_, err = os.Stat("todos.db")
	dbExists := !os.IsNotExist(err)

	// Open SQLite database
	db, err = sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create todos table if it doesn't exist
	if !dbExists {
		createTable := `
		CREATE TABLE todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			completed BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`
		_, err = db.Exec(createTable)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Todos table created")
	}

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database connection established")
}

func handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getTodos(w, r)
	case "POST":
		createTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTodo(w http.ResponseWriter, r *http.Request) {
	// Extract todo ID from URL
	id, err := strconv.Atoi(r.URL.Path[len("/todos/"):])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		getTodo(w, r, id)
	case "PUT":
		updateTodo(w, r, id)
	case "DELETE":
		deleteTodo(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTodos(w http.ResponseWriter, _ *http.Request) {
	// Query all todos from the database
	rows, err := db.Query("SELECT id, title, completed, created_at FROM todos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Process the results
	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		var createdAt string
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &createdAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todo.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		todos = append(todos, todo)
	}

	// Return todos as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func getTodo(w http.ResponseWriter, _ *http.Request, id int) {
	// Query specific todo from the database
	row := db.QueryRow("SELECT id, title, completed, created_at FROM todos WHERE id = ?", id)

	var todo Todo
	var createdAt string
	err := row.Scan(&todo.ID, &todo.Title, &todo.Completed, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	todo.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

	// Return todo as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert todo into the database
	result, err := db.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", todo.Title, todo.Completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly inserted todo
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the ID of the todo
	todo.ID = int(id)
	todo.CreatedAt = time.Now()

	// Return the newly created todo as JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	// Parse request body
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if todo exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM todos WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Update todo in the database
	_, err = db.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", todo.Title, todo.Completed, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the updated todo
	row := db.QueryRow("SELECT id, title, completed, created_at FROM todos WHERE id = ?", id)
	var createdAt string
	err = row.Scan(&todo.ID, &todo.Title, &todo.Completed, &createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	todo.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

	// Return the updated todo as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	// Check if todo exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM todos WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Delete todo from the database
	_, err = db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success message
	w.WriteHeader(http.StatusNoContent)
}