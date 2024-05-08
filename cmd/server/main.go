package main

import (
	"log"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Initialize Echo framework and templates
func initEcho() *echo.Echo {
	// Initialize templates
	controllers.InitTemplates()

	// Create a template registry for Echo
	e := echo.New()
	e.Renderer = controllers.TmplRegistry

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Serve static files
	e.Static("/static", "ui/static")

	return e
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
	e.GET("/", controllers.IndexHandler)
	e.GET("/sign-in", controllers.SignInHandler)
	e.GET("/create-account", controllers.CreateAccountHandler)
	e.GET("/create-question", controllers.CreateQuestionHandler)
	e.POST("/create-account-submit", controllers.CreateAccountSubmitHandler)
	e.POST("/login", controllers.LoginHandler)
	// Start server
	log.Println("Server is running at http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
