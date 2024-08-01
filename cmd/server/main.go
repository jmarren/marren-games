package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/awssdk"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/jmarren/marren-games/internal/routers"
	"github.com/jmarren/marren-games/internal/routers/restricted"
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
	envError := godotenv.Load()
	if envError != nil {
		log.Fatal("Error loading .env file")
	}
	//
	// AWS
	awssdk.InitAWS()

	// ---- Database
	dbConnectionError := db.InitDB()
	if dbConnectionError != nil {
		log.Fatalf("Failed to connect to the database: %v", dbConnectionError)
	} else {
		log.Print("DB connected successfully")
	}

	defer db.Sqlite.Close()
	// Initialize Echo
	e := initEcho()
	e.Use(middleware.Logger())

	pprofGroup := e.Group("/debug/pprof")
	pprofGroup.GET("/*", echo.WrapHandler(http.DefaultServeMux))
	// Unrestricted Routes
	unrestrictedRoutes := e.Group("")
	routers.UnrestrictedRoutes(unrestrictedRoutes)

	// Restricted Routes
	restrictedRoutes := e.Group("/auth")
	restricted.RestrictedRoutes(restrictedRoutes)
	//
	e2 := echo.New()

	e2.POST("/update-askers", db.UpdateAskers)
	//
	err := auth.CreateUserAndGameAtStartup()
	if err != nil {
		fmt.Println("error creating user at startup: ", err)
		panic(err)
	}

	// Channel to listen for termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Channel to wait for both servers to shut down
	done := make(chan struct{})

	// Register pprof routes
	// Start the main server on port 8080
	go func() {
		if err := e.Start(":8081"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	go func() {
		if err := e2.Start(":8082"); err != nil {
			e.Logger.Info("shutting down secure server")
		}
	}()

	// Wait for termination signal
	<-quit
	fmt.Println("Shutting down servers...")

	// Create a context with timeout to ensure graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down both servers
	go func() {
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
		if err := e2.Shutdown(ctx); err != nil {
			e2.Logger.Fatal(err)
		}
		close(done)
	}()

	// Wait for both servers to shut down
	<-done
	fmt.Println("Servers shut down gracefully")
}
