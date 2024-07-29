// db.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/tursodatabase/libsql-client-go/libsql"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/mattn/go-sqlite3"
)

var Sqlite *sql.DB

func InitDB() error {
	var err error
	var url string

	// get current working directory
	currentWorkingDir, _ := os.Getwd()

	// Read the init script
	initScript, err := os.ReadFile(currentWorkingDir + "/sql/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read init script: %v", err)
	}

	fmt.Println(os.Getenv("USE_DEV_SQLITE"))

	// Determine whether to use in-memory SQLite or the production database
	useDevSQLite := os.Getenv("USE_DEV_SQLITE")

	if useDevSQLite == "true" {
		fmt.Println("Using in-memory db")

		Sqlite, err = sql.Open("libsql", "file:memory:")
		if err != nil {
			return fmt.Errorf("failed to open in-memory db %s: %s", url, err)
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
	}

	if _, err := Sqlite.Exec(string(initScript)); err != nil {
		log.Fatalf("failed to execute init script: %v", err)
	}

	fmt.Println("Database initialized successfully")
	return nil
}

func UpdateAskers(c echo.Context) error {
	query := `
UPDATE current_askers
SET user_id = (
  SELECT (
    CASE
        WHEN (
          SELECT COUNT(*)
          FROM user_game_membership
          WHERE game_id = current_askers.game_id
            AND user_game_membership.user_id > current_askers.user_id
        ) > 0 THEN (
          SELECT user_id
          FROM user_game_membership
          WHERE current_askers.game_id = user_game_membership.game_id
            AND user_game_membership.user_id > current_askers.user_id
          ORDER BY user_game_membership.user_id
          LIMIT 1
          )
        ELSE (
          SELECT user_id
          FROM user_game_membership
          WHERE user_game_membership.game_id = current_askers.game_id
          ORDER BY user_game_membership.user_id
          LIMIT 1
        )
    END
  )
  FROM user_game_membership
);
`
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	// begin the Tx
	tx, err := Sqlite.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		cancel()
		return fmt.Errorf("error tx for updating askers: %v", err)
	}

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error updating askers: %v", err)
	}

	tx.Commit()
	return nil
}
