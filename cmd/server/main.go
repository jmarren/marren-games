package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	// AWS SDK
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String("ask-away"),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("first page results:")
	for _, object := range output.Contents {
		log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
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

	pprofGroup := e.Group("/debug/pprof")
	pprofGroup.GET("/*", echo.WrapHandler(http.DefaultServeMux))
	// Unrestricted Routes
	unrestrictedRoutes := e.Group("")
	routers.UnrestrictedRoutes(unrestrictedRoutes)

	// Restricted Routes
	restrictedRoutes := e.Group("/auth")
	routers.RestrictedRoutes(restrictedRoutes)

	// // Register pprof routes

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
