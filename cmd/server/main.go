package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplateRegistry struct {
	templates *template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Load all HTML files for templates, including the layout (index.html)
	t := &TemplateRegistry{
		templates: template.Must(template.ParseGlob(dir + "/ui/templates/*.html")),
	}

	log.Print("--------- Loaded Templates ----------")

	for _, tmpl := range t.templates.Templates() {
		log.Println(tmpl.Name())
	}

	log.Print("=====================================")

	e := echo.New()
	e.Renderer = t

	e.Use(middleware.Logger())

	// Serve static files
	e.Static("/static", "ui/static")

	// Home page route
	e.GET("/", func(c echo.Context) error {
		// Use the main index template with the home content template
		data := map[string]interface{}{
			"Content": "sign-in", // Home content template name
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	// Create Account route
	e.GET("/create-account", func(c echo.Context) error {
		data := map[string]interface{}{
			"Content": "create-account", // Create Account content template name
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	// Sign-In route
	e.GET("/sign-in", func(c echo.Context) error {
		data := map[string]interface{}{
			"Content": "sign-in", // Sign-In content template name
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	// Create Question route
	e.GET("/create-question", func(c echo.Context) error {
		data := map[string]interface{}{
			"Content": "create-question", // Create Question content template name
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

// package main
//
// import (
// 	"html/template"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
//
// 	// "github.com/jmarren/marren-games/internal/render"
// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// )
//
// // type User struct {
// // 	ID   int    `json:"id"`
// // 	Name string `json:"name"`
// // }
//
// type TemplateRegistry struct {
// 	templates *template.Template
// }
//
// func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
// 	return t.templates.ExecuteTemplate(w, name, data)
// }
//
// func main() {
// 	dir, err := os.Getwd()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	dir += "/ui/templates/*"
// 	// t := &TemplateRegistry{
// 	// 	templates: template.Must(template.ParseFiles(dir + "index.html")),
// 	// }
//
// 	t := &TemplateRegistry{
// 		templates: template.Must(template.ParseGlob(dir)),
// 	}
//
// 	log.Print("---------  t ----------")
//
// 	for _, tmpl := range t.templates.Templates() {
// 		log.Println(tmpl.Name())
// 	}
//
// 	log.Print("======================")
//
// 	e := echo.New()
// 	e.Renderer = t
//
// 	e.Use(middleware.Logger())
//
// 	// Serve static files
// 	e.Static("/static", "ui/static")
//
// 	e.GET("/", func(c echo.Context) error {
// 		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
// 			"content": "",
// 		})
// 	})
//
// 	e.GET("/create-account", func(c echo.Context) error {
// 		// define the data for the templates
// 		data := map[string]interface{}{
// 			"content": "create-account",
// 		}
//
// 		// execute the index.html template with the create-account template as the "content" template
// 		return c.Render(http.StatusOK, "index.html", data)
// 	})
//
// 	e.GET("/sign-in", func(c echo.Context) error {
// 		// Define the data for the templates
// 		data := map[string]interface{}{
// 			"content": "sign-in",
// 		}
//
// 		// Execute the index.html template with the sign-in template as the "content" template
// 		return c.Render(http.StatusOK, "index.html", data)
// 	})
//
// 	e.GET("/create-question", func(c echo.Context) error {
// 		// Define the data for the templates
// 		data := map[string]interface{}{
// 			"content": "create-question",
// 		}
//
// 		// Execute the index.html template with the create-question template as the "content" template
// 		return c.Render(http.StatusOK, "index.html", data)
// 	})
//
// 	// e.GET("/create-account", func(c echo.Context) error {
// 	// 	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
// 	// 		"content": "create-account",
// 	// 	})
// 	// })
//
// 	e.Logger.Fatal(e.Start(":8080"))
// }
//
//
//
//
//
//
//
//
///// First Working Version //////// =======================
// func main() {
// 	dir, err := os.Getwd()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	envError := godotenv.Load()
// 	if envError != nil {
// 		log.Fatal("Error loading .env file")
// 	}
//
// 	db.InitDB()
//
// 	e := echo.New()
//
// 	e.Use(middleware.Logger())
//
// 	// Serve static files
// 	e.Static("/static", "ui/static")
//
// 	e.GET("/", func(e echo.Context) error {
// 		// parse the template file in the current working directory
// 		tpl, err := template.ParseFiles(dir + "/ui/templates/index.html")
// 		if err != nil {
// 			fmt.Println(err)
// 			log.Fatal("Error parsing template file")
// 		}
//
// 		executionErr := tpl.Execute(e.Response().Writer, nil)
// 		if executionErr != nil {
// 			panic(executionErr)
// 		}
//
// 		return executionErr
// 	})
//
// 	e.GET("/create-account", func(e echo.Context) error {
// 		tpl, err := template.ParseFiles(dir+"/ui/templates/index.html", dir+"/ui/templates/blocks/create-account.html")
// 		if err != nil {
// 			fmt.Println(err)
// 			panic(err)
// 		}
//
// 		executionErr := tpl.Execute(e.Response().Writer, nil)
// 		if executionErr != nil {
// 			panic(executionErr)
// 		}
// 		return executionErr
// 	})
//
// 	e.Logger.Fatal(e.Start(":8080"))
// }
