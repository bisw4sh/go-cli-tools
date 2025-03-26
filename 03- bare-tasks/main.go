package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	Id    int
	Title string
}

var db *sql.DB

func initDB() {
	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS todos (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	title TEXT
	);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Error creating table: %q: %s\n", err, sqlStmt)
	}
}

// Add a new todo
func addTodo(title string) {
	_, err := db.Exec("INSERT INTO todos (title) VALUES (?)", title)
	if err != nil {
		log.Println("Failed to add todo:", err)
	} else {
		fmt.Println("Todo added!")
	}
}

// Show todos with pagination
func showTodos(offset int) int {
	rows, err := db.Query("SELECT id, title FROM todos LIMIT 10 OFFSET ?", offset)
	if err != nil {
		log.Println("Failed to fetch todos:", err)
		return offset
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.Id, &todo.Title)
		if err != nil {
			log.Println("Error scanning todo:", err)
			continue
		}
		fmt.Printf("%d. %s\n", todo.Id, todo.Title)
		count++
	}
	return count
}

// Find todo by id or pattern
func findTodos(query string) {
	rows, err := db.Query("SELECT id, title FROM todos WHERE title LIKE ? OR id = ?", "%"+query+"%", query)
	if err != nil {
		log.Println("Failed to find todos:", err)
		return
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.Id, &todo.Title)
		if err != nil {
			log.Println("Error scanning todo:", err)
			continue
		}
		fmt.Printf("%d. %s\n", todo.Id, todo.Title)
		found = true
	}

	if !found {
		fmt.Println("No matching todos found.")
	}
}

func main() {
	initDB()
	reader := bufio.NewReader(os.Stdin)
	offset := 0

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch {
		case strings.HasPrefix(input, "add "):
			title := strings.TrimPrefix(input, "add ")
			addTodo(title)

		case input == "show":
			count := showTodos(offset)
			if count == 0 {
				fmt.Println("No more todos.")
			} else {
				fmt.Println("Press Enter to see more or 'q' to go back.")
				for {
					cont, _ := reader.ReadString('\n')
					cont = strings.TrimSpace(cont)
					if cont == "q" {
						if offset >= 10 {
							offset -= 10
						}
						break
					} else if cont == "" {
						offset += 10
						count = showTodos(offset)
						if count == 0 {
							fmt.Println("No more todos.")
							break
						}
					} else {
						break
					}
				}
			}

		case strings.HasPrefix(input, "find "):
			query := strings.TrimPrefix(input, "find ")
			findTodos(query)

		case input == "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Use 'add', 'show', 'find', or 'exit'.")
		}
	}
}
