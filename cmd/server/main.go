package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/jmarren/marren-games/internal/awssdk"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/jmarren/marren-games/internal/routers"
	"github.com/jmarren/marren-games/internal/routers/queryTests"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// AWS
	awssdk.InitAWS()

	// ---- Database
	dbConnectionError := db.InitDB()
	if dbConnectionError != nil {
		log.Fatalf("Failed to connect to the database: %v", dbConnectionError)
	} else {
		log.Print("DB connected successfully")
	}

	// Initialize Echo
	e := initEcho()
	e.Use(middleware.Logger())

	queryTest := e.Group("/query")

	queryTests.QueryTests(queryTest)
	// ---- Middlewares
	// e.Use(fmt.Println(e.Logger()))
	// e.Use(e.Logger)
	// e.Use(middleware.Recover())

	// ---- Routes

	// Route for learning about go templates
	// e.GET("/learn-go-templates", controllers.Render())

	pprofGroup := e.Group("/debug/pprof")
	pprofGroup.GET("/*", echo.WrapHandler(http.DefaultServeMux))
	// Unrestricted Routes
	unrestrictedRoutes := e.Group("")
	routers.UnrestrictedRoutes(unrestrictedRoutes)

	// Restricted Routes
	restrictedRoutes := e.Group("/auth")
	routers.RestrictedRoutes(restrictedRoutes)

	// Register pprof routes
	// Start the main server on port 8080
	go func() {
		if err := e.Start(":8080"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// Set up a separate HTTP server for pprof on port 1323
	go func() {
		log.Println("Starting pprof server on :1323")
		if err := http.ListenAndServe(":1323", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for the application to terminate
	select {}
}
