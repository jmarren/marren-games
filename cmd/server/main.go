package main

import (
	"log"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	// echoprometheus  "github.com/labstack/echo-contrib"
	"github.com/labstack/echo/v4/middleware"
)

// Initialize Echo framework and templates
func initEcho() *echo.Echo {
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
	// Load environment variables
	envError := godotenv.Load()
	if envError != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the database
	dbConnectionError := db.InitDB()
	if dbConnectionError != nil {
		log.Fatalf("Failed to connect to the database: %v", dbConnectionError)
	} else {
		log.Print("DB connected successfully")
	}

	// Initialize Echo
	e := initEcho()

	// Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Routes

	// Unrestricted Routes
	e.GET("/", controllers.IndexHandler)
	e.GET("/sign-in", controllers.SignInHandler)
	e.GET("/create-account", controllers.CreateAccountHandler)
	e.POST("/login", controllers.LoginHandler)
	e.POST("/create-account-submit", controllers.CreateAccountSubmitHandler)

	// Restricted Routes
	r := e.Group("/auth")

	jwtConfig := echojwt.Config{
		SigningKey:  []byte("secret"),
		TokenLookup: "cookie:auth",
	}

	r.Use(echojwt.WithConfig(jwtConfig))

	r.GET("/index", controllers.IndexHandler)
	r.GET("/test", func(c echo.Context) error {
		return c.String(200, "You are authenticated")
	})
	// Start server
	log.Println("Server is running at http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
