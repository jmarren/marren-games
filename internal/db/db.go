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

var Sqlite *sql.DB

func InitDB() error {
	var err error
	var url string

	useDevSQLite := os.Getenv("USE_DEV_SQLITE")
	currentWorkingDir, _ := os.Getwd()
	log.Println("Current working directory: ", currentWorkingDir)
	initScript, err := os.ReadFile(currentWorkingDir + "/internal/db/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read init script: %v", err)
	}

	if useDevSQLite == "true" {
		log.Println("Using in-memory db")

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

	log.Println("Database initialized successfully")

	return nil
}

// package db
//
// import (
// 	"database/sql"
// 	"fmt"
// 	"os"
//
// 	_ "github.com/tursodatabase/libsql-client-go/libsql"
// )
//
// type User struct {
// 	ID   int
// 	Name string
// }
//
// var db *sql.DB
//
// func InitDB() error {
// 	databaseURL := os.Getenv("TURSO_DATABASE_URL")
// 	authToken := os.Getenv("TURSO_AUTH_TOKEN")
// 	url := fmt.Sprintf("%s?authToken=%s", databaseURL, authToken)
//
// 	fmt.Println("----- URL : ", url)
// 	var err error
// 	db, err = sql.Open("libsql", url)
// 	if err != nil {
// 		return fmt.Errorf("failed to open db %s: %s", url, err)
// 	}
//
// 	return nil
// }
//
// func GetUsers() ([]User, error) {
// 	rows, err := db.Query("SELECT * FROM users")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to execute query: %v", err)
// 	}
// 	defer rows.Close()
//
// 	var users []User
//
// 	for rows.Next() {
// 		var user User
//
// 		if err := rows.Scan(&user.ID, &user.Name); err != nil {
// 			return nil, fmt.Errorf("error scanning row: %v", err)
// 		}
//
// 		users = append(users, user)
// 	}
//
// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error during rows iteration: %v", err)
// 	}
//
// 	return users, nil
// }
//
// func queryUsers(db *sql.DB) {
// 	rows, err := db.Query("SELECT * FROM users")
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "failed to execute query: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer rows.Close()
//
// 	var users []User
//
// 	for rows.Next() {
// 		var user User
//
// 		if err := rows.Scan(&user.ID, &user.Name); err != nil {
// 			fmt.Println("Error scanning row:", err)
// 			return
// 		}
//
// 		users = append(users, user)
// 		fmt.Println(user.ID, user.Name)
// 	}
//
// 	if err := rows.Err(); err != nil {
// 		fmt.Println("Error during rows iteration:", err)
// 	}
// }
//
// func main() {
// 	databaseURL := os.Getenv("TURSO_DATABASE_URL")
// 	authToken := os.Getenv("TURSO_AUTH_TOKEN")
// 	url := fmt.Sprintf("%s?authToken=%s", databaseURL, authToken)
//
// 	fmt.Println(url)
//
// 	db, err := sql.Open("libsql", url)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
// 		os.Exit(1)
// 	}
//
// 	defer db.Close()
// }
