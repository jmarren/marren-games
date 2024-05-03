package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jmarren/marren-games/internal/db"
	"github.com/jmarren/marren-games/internal/handler/render"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getUsers(c echo.Context) error {
	users, err := db.GetUsers()
	if err != nil {
		fmt.Println("error getting users: ", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db.InitDB()

	e := echo.New()
	e.Renderer = render.NewTemplates()
	e.Use(middleware.Logger())

	// Serve static files
	e.Static("/static", "ui/static")

	e.GET("/", func(c echo.Context) error {
		data := &render.PageData{
			Options: []string{"Mom", "Dad", "Tom", "Anna", "Megan", "Robby", "Allie", "Kristin", "Kevin", "John"},
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	e.GET("/home", func(c echo.Context) error {
		data := &render.PageData{
			Options: []string{"Mom", "Dad", "Tom", "Anna", "Megan", "Robby", "Allie", "Kristin", "Kevin", "John"},
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	e.GET("/users", getUsers)

	e.Logger.Fatal(e.Start(":8080"))
}
