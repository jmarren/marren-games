// Handles core authentication logic
package auth

import (
	"github.com/jmarren/marren-games/internal/db"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a hashed password from a plain string.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash checks if the provided password matches the stored hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func AuthenticateUser(username, password string) (bool, error) {
	// Get the hashed password from the database
	hashedPassword, err := db.GetUserPasswordHash(username, password)
	if err != nil {
		return false, err
	}

	// Compare the hashed password with the plaintext password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// bcrypt returns an error if the hashes don't match
		return false, nil
	}

	// The hashes match
	return true, nil
}

// RegisterUser creates a new user in the database.
