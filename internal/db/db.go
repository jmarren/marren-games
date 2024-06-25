// db.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/mattn/go-sqlite3"
)

var (
	Sqlite      *sql.DB
	WithQueries *WithQueriesMap
)

func InitDB() error {
	var err error
	var url string

	// get current working directory
	currentWorkingDir, _ := os.Getwd()

	// Read the init script
	initScript, err := os.ReadFile(currentWorkingDir + "/internal/sql/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read init script: %v", err)
	}

	// Determine whether to use in-memory SQLite or the production database
	useDevSQLite := os.Getenv("USE_DEV_SQLITE")

	if useDevSQLite == "true" {
		fmt.Println("Using in-memory db")

		Sqlite, err = sql.Open("libsql", "file:memory:")
		if err != nil {
			return fmt.Errorf("failed to open in-memory db %s: %s", url, err)
		}

		if _, err := Sqlite.Exec(string(initScript)); err != nil {
			log.Fatalf("failed to execute init script: %v", err)
		}
		// Specify the backup file path
		backupFilePath := "backup.db"

		// Check if the backup file exists and remove it
		if _, err := os.Stat(backupFilePath); err == nil {
			if err := os.Remove(backupFilePath); err != nil {
				log.Fatalf("Failed to remove existing backup file %s: %v", backupFilePath, err)
			}
		}
		_, err := Sqlite.Exec("VACUUM INTO 'backup.db';")
		if err != nil {
			log.Fatalf("Failed to backup in-memory database: %v", err)
		}
		log.Println("Database backed up successfully")
	} else {

		databaseURL := os.Getenv("TURSO_DATABASE_URL")
		authToken := os.Getenv("TURSO_AUTH_TOKEN")
		url := fmt.Sprintf("%s?authToken=%s", databaseURL, authToken)

		var err error
		Sqlite, err = sql.Open("libsql", url)
		if err != nil {
			return fmt.Errorf("failed to open db %s: %s", url, err)
		}

		if _, err := Sqlite.Exec(string(initScript)); err != nil {
			log.Fatalf("failed to execute init script: %v", err)
		}
	}

	fmt.Println("Database initialized successfully")

	// Get the WithQueriesMap
	fmt.Println("Getting WithQueriesMap")

	WithQueries = CreateWithQueriesMap()

	fmt.Println("WithQueriesMap created successfully")

	return nil
}
