package db // import "github.com/jmarren/marren-games/internal/db"

import (
	"fmt"
	"os"
)

func AddUser(username, hashedPassword, email string) error {
	// Ensure the database connection is initialized
	if db == nil {
		fmt.Println("db not connected")
	}

	fmt.Printf("username: %s \n hashedPassword: %s \n email: %s", username, hashedPassword, email)
	_, err := db.Exec(`INSERT INTO users (username, password_hash, email) VALUES (?, ?, ?)`, username, string(hashedPassword), email)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute query: %v\n", err)
		os.Exit(1)
	}

	return nil
}

// func GetCurrentAnswer(username) error {
// 	err := db.Exec(`SELECT answer_text FROM answers WHERE user = `)
// }

func GetUserPasswordHash(username string) (string, error) {
	var hashedPassword string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func SetCurrentQuestion(askerId int, questionText string) {
	_, err := db.Exec(`INSERT INTO questions (asker_id, question_text) VALUES (?, ?)`, askerId, questionText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute query: %v\n", err)
		os.Exit(1)
	}
}

func GetCurrentQuestion(date_created string) (string, error) {
	var questionText string
	err := db.QueryRow("SELECT question_text FROM questions WHERE date_created = ", date_created).Scan(&questionText)
	if err != nil {
		return "", err
	}
	return questionText, nil
}
