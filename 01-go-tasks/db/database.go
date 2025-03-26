package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens a database connection and initializes the schema if needed
func InitDB(dbPath string) (*sql.DB, error) {
	// Check if the database file exists
	_, err := os.Stat(dbPath)
	dbExists := !os.IsNotExist(err)

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create todos table if it doesn't exist
	if !dbExists {
		err = createSchema(db)
		if err != nil {
			return nil, err
		}
		fmt.Println("Todos table created")
	}

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Database connection established")

	return db, nil
}

// createSchema creates the necessary tables
func createSchema(db *sql.DB) error {
	createTable := `
	CREATE TABLE todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(createTable)
	return err
}