package db // import "github.com/jmarren/marren-games/internal/db"

import (
	"database/sql"
)

func RegisterUser(db *sql.DB, username, hashedPassword, email string) error {
	// Insert user into database
	_, err := db.Exec("INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)", username, hashedPassword, email)
	if err != nil {
		return err
	}

	return nil
}

func GetUserPasswordHash(db *sql.DB, username string) (string, error) {
	var hashedPassword string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

// package db // import "github.com/jmarren/marren-games/internal/db"
//
// import (
// 	"database/sql"
//
// 	"golang.org/x/crypto/bcrypt"
// )
//
// func RegisterUser(db *sql.DB, username, password, email string) error {
// 	// Hash the password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Insert user into database
// 	_, err = db.Exec("INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)", username, string(hashedPassword), email)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
//
// func VerifyUser(db *sql.DB, username, password string) (bool, error) {
//   var hashedPassword string
//   err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&hashedPassword)
//   if err != nil {
//     return false, err
//   }
//   err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
//   if err != nil {
//     return false, nil
//   }
//   return true, nil
// }
//
//
//
