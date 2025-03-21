package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-task/models"
)

// TodoHandler manages HTTP requests for todos
type TodoHandler struct {
	db *sql.DB
}

// NewTodoHandler creates a new TodoHandler with the given database
func NewTodoHandler(db *sql.DB) *TodoHandler {
	return &TodoHandler{db: db}
}

// HandleTodos handles requests to /todos endpoint
func (h *TodoHandler) HandleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getTodos(w, r)
	case http.MethodPost:
		h.createTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTodo handles requests to /todos/{id} endpoint
func (h *TodoHandler) HandleTodo(w http.ResponseWriter, r *http.Request) {
	// Extract todo ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/todos/")
	if path == "" {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTodo(w, r, id)
	case http.MethodPut:
		h.updateTodo(w, r, id)
	case http.MethodDelete:
		h.deleteTodo(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TodoHandler) getTodos(w http.ResponseWriter, r *http.Request) {
	// Query all todos from the database
	rows, err := h.db.Query("SELECT id, title, completed, created_at FROM todos")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Process the results
	todos := []models.Todo{}
	for rows.Next() {
		var todo models.Todo
		var createdAt string
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &createdAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todo.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		todos = append(todos, todo)
	}

	// Return todos as JSON response
	json.NewEncoder(w).Encode(todos)
}

func (h *TodoHandler) getTodo(w http.ResponseWriter, r *http.Request, id int) {
	// Query specific todo from the database
	row := h.db.QueryRow("SELECT id, title, completed, created_at FROM todos WHERE id = ?", id)

	var todo models.Todo
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
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) createTodo(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var todo models.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert todo into the database
	result, err := h.db.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", todo.Title, todo.Completed)
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) updateTodo(w http.ResponseWriter, r *http.Request, id int) {
	// Parse request body
	var todo models.Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if todo exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM todos WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Update todo in the database
	_, err = h.db.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", todo.Title, todo.Completed, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the updated todo
	row := h.db.QueryRow("SELECT id, title, completed, created_at FROM todos WHERE id = ?", id)
	var createdAt string
	err = row.Scan(&todo.ID, &todo.Title, &todo.Completed, &createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	todo.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

	// Return the updated todo as JSON response
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
	// Check if todo exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM todos WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Delete todo from the database
	_, err = h.db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success message
	w.WriteHeader(http.StatusNoContent)
}
