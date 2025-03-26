package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Todo struct {
	ID        int    `json:"id"`
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "todos.db")
	if err != nil {
		log.Fatal(err)
	}

	query := `CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0
	)`
	if _, err = db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func createTodoHandler(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO todos (task, completed) VALUES (?, ?)", todo.Task, todo.Completed)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	todo.ID = int(id)
	json.NewEncoder(w).Encode(todo)
}

func getTodosHandler(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query("SELECT id, task, completed FROM todos")
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Task, &t.Completed); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	json.NewEncoder(w).Encode(todos)
}

func updateTodoHandler(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE todos SET task = ?, completed = ? WHERE id = ?", todo.Task, todo.Completed, todo.ID)
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(todo)
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM todos WHERE id = ?", todo.ID)
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Todo deleted"})
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTodosHandler(w, r)
		case http.MethodPost:
			createTodoHandler(w, r)
		case http.MethodPut:
			updateTodoHandler(w, r)
		case http.MethodDelete:
			deleteTodoHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
