package main

import (
	"log"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/jmarren/marren-games/internal/routers"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	// echoprometheus  "github.com/labstack/echo-contrib"
	// "github.com/labstack/echo/v4/middleware"
)

// Initialize Echo framework and templates
func initEcho() *echo.Echo {
	controllers.InitTemplates()

	// Create a template registry for Echo
	e := echo.New()
	e.Renderer = controllers.TmplRegistry

	// Middleware
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	// Serve static files
	e.Static("/static", "ui/static")

	return e
}

func main() {
	// ---- Env Variables
	envError := godotenv.Load()
	if envError != nil {
		log.Fatal("Error loading .env file")
	}

	// ---- Database
	dbConnectionError := db.InitDB()
	if dbConnectionError != nil {
		log.Fatalf("Failed to connect to the database: %v", dbConnectionError)
	} else {
		log.Print("DB connected successfully")
	}

	// Initialize Echo
	e := initEcho()

	queryTest := e.Group("/query")

	routers.QueryTests(queryTest)
	// ---- Middlewares
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	// ---- Routes

	// Route for learning about go templates
	// e.GET("/learn-go-templates", controllers.Render())

	// Unrestricted Routes
	unrestrictedRoutes := e.Group("")
	routers.UnrestrictedRoutes(unrestrictedRoutes)

	// Restricted Routes
	restrictedRoutes := e.Group("/auth")
	routers.RestrictedRoutes(restrictedRoutes)

	// Start server
	log.Println("Server is running at http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
