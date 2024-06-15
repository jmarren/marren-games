package controllers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/auth"
	"github.com/labstack/echo/v4"
)

// TemplateRegistry is a custom HTML template renderer for Echo framework
type TemplateRegistry struct {
	templates *template.Template
}

// Render implements the Echo Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var TmplRegistry *TemplateRegistry

type PageData struct {
	Title           string
	PartialTemplate string
	Data            interface{}
}

type CreateAccountData struct {
	Username string
	Email    string
	Error    error
}
type SignInData struct {
	Username string
}
type CreateAccountSuccessData struct {
	Username string
}

type ProfileData struct {
	Username      string
	GameCompleted bool
}

// Initialize templates
func InitTemplates() {
	// Determine the working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	basePath := "ui/templates/"
	partials := []string{
		"create-account.html",
		"sign-in.html",
		"create-question.html",
		"index.html",
		"create-account-success.html",
		"profile.html",
	}

	// Create a base layout template
	templates := template.New("base").Funcs(template.FuncMap{})

	// Parse base layout
	templates = template.Must(templates.ParseFiles(dir + "/" + basePath + "base.html"))

	// Parse all partial templates (blocks)
	for _, partial := range partials {
		templates = template.Must(templates.ParseFiles(dir + "/" + basePath + partial))
	}

	// Set up the global template registry
	TmplRegistry = &TemplateRegistry{templates: templates}
}

// Render full or partial templates based on the HX-Request header
func renderTemplate(c echo.Context, partialTemplate string, data interface{}) error {
	hx := c.Request().Header.Get("HX-Request") != ""

	if hx {
		// HTMX Request: Render only the partial content
		return c.Render(http.StatusOK, partialTemplate, data)
	} else {
		// Full Page Reload: Render the base layout with the partial content
		return c.Render(http.StatusOK, "base", PageData{
			Title:           "Marren Games",
			PartialTemplate: partialTemplate,
			Data:            data,
		})
	}
}

func IndexHandler(c echo.Context) error {
	return renderTemplate(c, "index", nil)
}

func SignInHandler(c echo.Context) error {
	return renderTemplate(c, "sign-in", nil)
}

func CreateAccountHandler(c echo.Context) error {
	return renderTemplate(c, "create-account", nil)
}

func CreateQuestionHandler(c echo.Context) error {
	return renderTemplate(c, "create-question", nil)
}

func CreateAccountSubmitHandler(c echo.Context) error {
	// log the form values
	log.Print("username:", c.FormValue("username"))
	registrationError := auth.RegisterUser(c.FormValue("username"), c.FormValue("password"), c.FormValue("email"))
	if registrationError == nil {
		return renderTemplate(c, "create-account-success", CreateAccountSuccessData{
			Username: c.FormValue("username"),
		})
	}
	data := CreateAccountData{
		Username: c.FormValue("username"),
		Email:    c.FormValue("email"),
		Error:    registrationError,
	}

	return renderTemplate(c, "create-account", data)
}

func LoginHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	jwt, err := auth.AuthenticateUser(username, password)
	if err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred during authentication.")
	}
	cookie := &http.Cookie{
		Name:    "auth",
		Value:   jwt,
		Expires: time.Now().Add(24 * time.Hour),
	}

	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/auth/profile")
}

func LogoutHandler(c echo.Context) error {
	cookie := &http.Cookie{
		Name:    "auth",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	}
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/")
}

func ProfileHandler(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*auth.JwtCustomClaims)
	username := claims.Username
	fmt.Println("------------ User: ", username, "--------------")
	return c.String(http.StatusOK, "Profile")
}

// type AnswerQuestionData struct {
// 	Question string
// 	Choices  []string
// }
//
// type CreateAccountData struct {
// 	Username string
// 	Email    string
// 	Success  bool
// }
//
// type RegistrationResult struct {
// 	Title    string
// 	Success  bool
// 	Username string
// }
//
// //	e.GET("/", func(c echo.Context) error {
// //	    return c.Render(http.StatusOK, "index.html", map[string]interface{}{
// //	        "Title": "Home Page",
// //	        "ContentTemplate": "home.html", // Assuming you have a home.html for the homepage
// //	    })
// //	})
// func RenderIndex(c echo.Context) error {
// 	return c.Render(http.StatusOK, "create-account", map[string]interface{}{
// 		"Title":           "Create Account",
// 		"ContentTemplate": "create-account.html",
// 	})
// }
//
// func RenderHome(c echo.Context) error {
// 	data := struct{}{}
// 	return c.Render(http.StatusOK, "create-account", data)
// }
//
// func RenderLoginResult(c echo.Context) error {
// 	username := c.FormValue("username")
// 	password := c.FormValue("password")
//
// 	isAuthenticated, err := auth.AuthenticateUser(username, password)
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, "An error occurred during authentication.")
// 	}
//
// 	if isAuthenticated {
// 		data := &render.PageData{
// 			Options: []string{"Mom", "Dad", "Tom", "Anna", "Megan", "Robby", "Allie", "Kristin", "Kevin", "John"},
// 		}
// 		return c.Render(http.StatusOK, "home.html", data)
// 	} else {
// 		return c.Render(http.StatusUnauthorized, "error.html", nil)
// 	}
// }
//
// func AnswerQuestionController(c echo.Context) error {
// 	data := AnswerQuestionData{
// 		Question: "What is your favorite color?",
// 		Choices:  []string{"Red", "Blue", "Green"},
// 	}
// 	return c.Render(http.StatusOK, "answer-question", data)
// }
//
// // func CreateAccountController(c echo.Context) error {
// // 	data := CreateAccountData{
// // 		Username: c.FormValue(name),
// // 		Email: c.FormValue(email),
// //     Success: true,
// // 	}
// // 	return c.Render(http.StatusOK, "create-account", data)
// // }
//
// func CreateAccountController(c echo.Context) error {
// 	// Create a new instance of CreateAccountData
// 	data := new(CreateAccountData)
//
// 	// Bind the form data to the CreateAccountData instance
// 	if err := c.Bind(data); err != nil {
// 		// If there's an error, respond with an error message
// 		return c.String(http.StatusBadRequest, "Invalid form data")
// 	}
//
// 	// If there's no error, render the create-account-success template
// 	return c.Render(http.StatusOK, "create-account-success", data)
// }
//
// // func RenderCreateAccountSuccessful(c echo.Context) error {
// // 	username := c.FormValue("username")
// // 	password := c.FormValue("password")
// // 	email := c.FormValue("email")
// // 	err := auth.RegisterUser(username, password, email)
// // 	if err != nil {
// // 		return c.String(http.StatusInternalServerError, "An error occurred during registration.")
// // 	}
// // 	data := &RegistrationResult{
// //     Title: "create-account",
// // 		Success:  true,
// // 		Username: username,
// // 	}
// // 	return c.Render(http.StatusOK, "index.html", data)
// // }
