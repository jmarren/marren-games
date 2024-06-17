package db // import "github.com/jmarren/marren-games/internal/db"

import (
	"fmt"
	"os"
)

func AddUser(username, hashedPassword, email string) error {
	// Ensure the database connection is initialized
	if Sqlite == nil {
		fmt.Println("db not connected")
	}

	fmt.Printf("username: %s \n hashedPassword: %s \n email: %s", username, hashedPassword, email)
	_, err := Sqlite.Exec(`INSERT INTO users (username, password_hash, email) VALUES (?, ?, ?)`, username, string(hashedPassword), email)
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
	err := Sqlite.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func SetCurrentQuestion(askerId int, questionText string) {
	_, err := Sqlite.Exec(`INSERT INTO questions (asker_id, question_text) VALUES (?, ?)`, askerId, questionText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute query: %v\n", err)
		os.Exit(1)
	}
}

func GetCurrentQuestion() (string, error) {
	var questionText string

	query := "SELECT question_text FROM questions WHERE date(date_created) = date(CURRENT_TIMESTAMP, '-6 hours')"
	err := Sqlite.QueryRow(query).Scan(&questionText)
	if err != nil {
		return "", err
	}
	return questionText, nil
}

func GetUserIdFromUsername(username string) (int, error) {
	var id int
	query := `SELECT id FROM users WHERE username = ?`
	err := Sqlite.QueryRow(query, username).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetUsernameFromUserId(id int) (string, error) {
	var username string
	query := `SELECT username FROM users WHERE id = ?`
	err := Sqlite.QueryRow(query, id).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func GetTodaysAnswerFromUserId(id int) (string, error) {
	var answerText string

	query := `SELECT a.answer_text
            FROM users u
            JOIN questions q ON u.id = q.asker_id 
            JOIN answers a ON q.id = a.question_id
            WHERE u.id = ?
            AND date(q.date_created) = date(CURRENT_TIMESTAMP, '-6 hours')
            `

	err := Sqlite.QueryRow(query, id).Scan(&answerText)
	if err != nil {
		return "", err
	}
	return answerText, nil
}

func QueryRowHandler(query string, args ...interface{}) string {
	var output string

	err := Sqlite.QueryRow(query, args...).Scan(&output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute query: %v\n", err)
	}
	return output
}

// func GetProfileData(username string) {
//
//   query := `SELECT id FROM users WHERE username = ?,
//               `
//
// }
