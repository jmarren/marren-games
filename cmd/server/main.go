package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("ui/templates/*.html")),
	}
}

type PageData struct {
	BlockColor string
}

func main() {
	e := echo.New()
	e.Renderer = NewTemplates()
	e.Use(middleware.Logger())

	// Serve static files
	e.Static("/static", "ui/static")

	e.GET("/", func(c echo.Context) error {
		data := &PageData{
			BlockColor: "#FF0000",
		}
		return c.Render(http.StatusOK, "blocks.html", data)
	})

	e.GET("/home", func(c echo.Context) error {
		data := &PageData{
			BlockColor: "#FF0000",
		}
		return c.Render(http.StatusOK, "blocks.html", data)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
