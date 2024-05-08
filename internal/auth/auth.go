// Handles core authentication logic
package auth

import (
	"errors"
	"net/mail"
	"slices"

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
	hashedPassword, err := db.GetUserPasswordHash(username)
	if err != nil {
		return false, err
	}

	// Compare the hashed password with the plain text password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// bcrypt returns an error if the hashes don't match
		return false, nil
	}

	// The hashes match
	return true, nil
}

// RegisterUser creates a new user in the database.
// It hashes the password before storing it.
func RegisterUser(username, password, email string) error {
	// First determine if the data is valid
	// Validate the username
	acceptableUsernames := []string{"admin", "John", "Kevin", "Anna", "Megan", "Tom", "Kristin", "Allie", "Robby", "Mom", "Dad"}

	if !slices.Contains(acceptableUsernames, username) {
		return errors.New("username is not allowed")
	}
	// Validate the password
	// This is a simple check, you may want to enforce more complex rules
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 16 {
		return errors.New("password must be less than 16 characters")
	}

	// Validate the email
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email")
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Insert user into database
	err = db.AddUser(username, hashedPassword, email)
	if err != nil {
		return err
	}

	return nil
}
