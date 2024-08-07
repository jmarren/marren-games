// Handles core authentication logic
package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/mail"
	"os"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"

	"golang.org/x/crypto/bcrypt"
)

type JwtCustomClaims struct {
	Username string `json:"username"`
	UserId   int    `json:"userId"`
	// UserPhotoVersion int    `json:"userPhotoVersion"`
	jwt.RegisteredClaims
}

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

func AuthenticateUser(username, password string) (string, error) {
	// Get the hashed password from the database
	hashedPassword, err := db.GetUserPasswordHash(username)
	if err != nil {
		return "", fmt.Errorf("error getting hased password from db: %v", err)
	}

	// Compare the hashed password with the plain text password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// bcrypt returns an error if the hashes don't match
		return "", err
	}

	userId, err := db.GetUserIdFromUsername(username)
	if err != nil {
		return "", err
	}

	// userPhotoVersion, err := db.GetUserPhotoVersionFromUsername(username)
	// if err != nil {
	// 	return "", err
	// }

	log.Println("User authenticated successfully")

	// Set custom claims
	claims := &JwtCustomClaims{
		username,
		userId,
		// userPhotoVersion,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	secret := os.Getenv("JWTSECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	log.Println("token: ", t)

	// The hashes match
	return t, nil
}

func CreateUserAndGameAtStartup() error {
	myUsername := os.Getenv("MY_USERNAME")
	myPassword := os.Getenv("MY_PASSWORD")
	myEmail := os.Getenv("MY_EMAIL")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()
	// begin the Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	defer tx.Rollback()

	if err != nil {
		cancel()
		return fmt.Errorf("error tx for adding user at startup: %v", err)
	}

	hashedPassword, err := HashPassword(myPassword)
	if err != nil {
		return fmt.Errorf("error hashing password CreateUserOnStartup(): %v", err)
	}

	query := `
    INSERT OR IGNORE INTO users (id, username, email, password_hash)
    VALUES (1, ?, ?, ?);
  `

	// Insert user into database
	_, err = tx.ExecContext(ctx, query, myUsername, myEmail, hashedPassword)
	if err != nil {
		fmt.Println("error adding user: ", err)
		return err
	}
	query = `
	  INSERT OR IGNORE INTO games (name, creator_id)
	  SELECT "All Users", id
	  FROM users
	  WHERE username = "John";
	  `

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while creating all users game: %v", err)
	}

	return tx.Commit()
}

// RegisterUser creates a new user in the database.
// It hashes the password before storing it.
func RegisterUser(username, password, email string) (string, error) {
	// First determine if the data is valid
	// Validate the username
	acceptableUsernames := []string{"admin", "John", "Kevin", "Anna", "Megan", "Tom", "Kristin", "Allie", "Robby", "Mom", "Dad"}

	if !slices.Contains(acceptableUsernames, username) {
		return "", errors.New("username is not allowed")
	}
	// Validate the password
	// This is a simple check, you may want to enforce more complex rules
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	if len(password) > 16 {
		return "", errors.New("password must be less than 16 characters")
	}

	// Validate the email
	_, err := mail.ParseAddress(email)
	if err != nil {
		return "", errors.New("invalid email")
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	// Insert user into database
	result, err := db.AddUser(username, hashedPassword, email)
	if err != nil {
		fmt.Println("error adding user: ", err)
		return "", err
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	var userId int
	// Ensure the value fits in an int to avoid overflow
	if insertId < math.MinInt || insertId > math.MaxInt {
		fmt.Println("Error: int64 value is out of int range")
	} else {
		userId = int(insertId)
		fmt.Printf("Converted value: %d\n", userId)
	}

	// Set custom claims
	claims := &JwtCustomClaims{
		username,
		userId,
		// 0,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("JWTSECRET")))
	if err != nil {
		return "", err
	}
	log.Println("token: ", t)

	return t, nil
}

type ClaimsType string

const (
	Username ClaimsType = "Username"
	UserId   ClaimsType = "UserId"
	// UserPhotoVersion ClaimsType = "UserPhotoVersion"
)

func GetFromClaims(item ClaimsType, c echo.Context) interface{} {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil
	}
	claims, ok := user.Claims.(*JwtCustomClaims)
	if !ok {
		return nil
	}

	switch item {
	case Username:
		return claims.Username
	case UserId:
		return claims.UserId
	// case UserPhotoVersion:
	// 	return claims.UserPhotoVersion
	default:
		return ""
	}
}
