package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRegistry is a custom HTML template renderer for Echo framework
type TemplateRegistry struct {
	templates *template.Template
}

// Render implements the Echo Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var tmplRegistry *TemplateRegistry

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
	Email    string
}
type CreateAccountSuccessData struct {
	Username string
}

// Initialize templates
func initTemplates() {
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
	tmplRegistry = &TemplateRegistry{templates: templates}
}

// Initialize Echo framework and templates
func initEcho() *echo.Echo {
	// Initialize templates
	initTemplates()

	// Create a template registry for Echo
	e := echo.New()
	e.Renderer = tmplRegistry

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Serve static files
	e.Static("/static", "ui/static")

	return e
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

func indexHandler(c echo.Context) error {
	return renderTemplate(c, "index", nil)
}

func signInHandler(c echo.Context) error {
	return renderTemplate(c, "sign-in", nil)
}

func createAccountHandler(c echo.Context) error {
	return renderTemplate(c, "create-account", nil)
}

func createQuestionHandler(c echo.Context) error {
	return renderTemplate(c, "create-question", nil)
}

func createAccountSubmitHandler(c echo.Context) error {
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

func loginHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	email := c.FormValue("email")

	data := SignInData{
		Username: username,
		Email:    email,
	}
	authResult, err := auth.AuthenticateUser(username, password)
	if err != nil {
		log.Fatal(err)
	}
	if authResult {
		return renderTemplate(c, "index", nil)
	}

	return renderTemplate(c, "sign-in", data)
}

func main() {
	envError := godotenv.Load()
	if envError != nil {
		log.Fatal("Error loading .env file")
	}
	dbConnectionError := db.InitDB()
	if dbConnectionError != nil {
		log.Fatalf("Failed to connect to the database: %v", dbConnectionError)
	} else {
		log.Print("DB connected successfully")
	}
	//
	e := initEcho()

	// Routes
	e.GET("/", indexHandler)
	e.GET("/sign-in", signInHandler)
	e.GET("/create-account", createAccountHandler)
	e.GET("/create-question", createQuestionHandler)
	e.POST("/create-account-submit", createAccountSubmitHandler)
	e.POST("/login", loginHandler)
	// Start server
	log.Println("Server is running at http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
