package controllers

import (
	"net/http"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/render"
	"github.com/labstack/echo/v4"
)

type AnswerQuestionData struct {
	Question string
	Choices  []string
}

type CreateAccountData struct {
	Username string
	Email    string
	Success  bool
}

type RegistrationResult struct {
	Title    string
	Success  bool
	Username string
}

//	e.GET("/", func(c echo.Context) error {
//	    return c.Render(http.StatusOK, "index.html", map[string]interface{}{
//	        "Title": "Home Page",
//	        "ContentTemplate": "home.html", // Assuming you have a home.html for the homepage
//	    })
//	})
func RenderIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "create-account", map[string]interface{}{
		"Title":           "Create Account",
		"ContentTemplate": "create-account.html",
	})
}

func RenderHome(c echo.Context) error {
	data := struct{}{}
	return c.Render(http.StatusOK, "create-account", data)
}

func RenderLoginResult(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	isAuthenticated, err := auth.AuthenticateUser(username, password)
	if err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred during authentication.")
	}

	if isAuthenticated {
		data := &render.PageData{
			Options: []string{"Mom", "Dad", "Tom", "Anna", "Megan", "Robby", "Allie", "Kristin", "Kevin", "John"},
		}
		return c.Render(http.StatusOK, "home.html", data)
	} else {
		return c.Render(http.StatusUnauthorized, "error.html", nil)
	}
}

func AnswerQuestionController(c echo.Context) error {
	data := AnswerQuestionData{
		Question: "What is your favorite color?",
		Choices:  []string{"Red", "Blue", "Green"},
	}
	return c.Render(http.StatusOK, "answer-question", data)
}

// func CreateAccountController(c echo.Context) error {
// 	data := CreateAccountData{
// 		Username: c.FormValue(name),
// 		Email: c.FormValue(email),
//     Success: true,
// 	}
// 	return c.Render(http.StatusOK, "create-account", data)
// }

func CreateAccountController(c echo.Context) error {
	// Create a new instance of CreateAccountData
	data := new(CreateAccountData)

	// Bind the form data to the CreateAccountData instance
	if err := c.Bind(data); err != nil {
		// If there's an error, respond with an error message
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// If there's no error, render the create-account-success template
	return c.Render(http.StatusOK, "create-account-success", data)
}

// func RenderCreateAccountSuccessful(c echo.Context) error {
// 	username := c.FormValue("username")
// 	password := c.FormValue("password")
// 	email := c.FormValue("email")
// 	err := auth.RegisterUser(username, password, email)
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, "An error occurred during registration.")
// 	}
// 	data := &RegistrationResult{
//     Title: "create-account",
// 		Success:  true,
// 		Username: username,
// 	}
// 	return c.Render(http.StatusOK, "index.html", data)
// }
