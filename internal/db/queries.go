package db // import "github.com/jmarren/marren-games/internal/db"

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(db *sql.DB, username, password, email string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert user into database
	_, err = db.Exec("INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)", username, string(hashedPassword), email)
	if err != nil {
		return err
	}

	return nil
}
