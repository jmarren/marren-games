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
	return auth.RegisterUser(c.FormValue("username"), c.FormValue("password"), c.FormValue("email"))
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

	// Start server
	log.Println("Server is running at http://localhost:8080")
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
// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// )
//
// // TemplateRegistry is a custom HTML template renderer for Echo framework
// type TemplateRegistry struct {
// 	templates map[string]*template.Template
// }
//
// func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
// 	tmpl, ok := t.templates[name]
// 	if !ok {
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Template not found: "+name)
// 	}
// 	return tmpl.ExecuteTemplate(w, "index.html", data)
// }
//
// // Initialize templates
// func initTemplates() {
//
// 	// Determine the working directory
// 	dir, err := os.Getwd()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// }
//
//
//
// // Initialize Echo framework and templates
// func initEcho() *echo.Echo {
// 	// Initialize templates
// 	initTemplates()
//
// 	// Create a template registry for Echo
// 	e := echo.New()
// 	e.Renderer = &TemplateRegistry{
// 		templates: fullPages,
// 	}
//
// 	// Middleware
// 	e.Use(middleware.Logger())
// 	e.Use(middleware.Recover())
//
// 	// Serve static files
// 	e.Static("/static", "static")
//
// 	return e
// }
//
// // Render full or partial templates based on the HX-Request header
// func renderTemplate(c echo.Context, tmpl string, data interface{}) error {
// 	hx := c.Request().Header.Get("HX-Request") != ""
// 	var t *template.Template
// 	var ok bool
//
// 	log.Print("HX-Request ? ---", hx)
// 	if hx {
// 		t, ok = partialPages[tmpl]
// 	} else {
// 		t, ok = fullPages[tmpl]
// 	}
//
// 	if !ok {
// 		log.Print("Template not found: ", tmpl)
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Template not found: "+tmpl)
// 	}
//
// 	if hx {
// 		return t.Execute(c.Response().Writer, data)
// 	}
// 	return c.Render(http.StatusOK, tmpl, data)
// }
//
// func indexHandler(c echo.Context) error {
// 	return renderTemplate(c, "index", nil)
// }
//
// func signInHandler(c echo.Context) error {
// 	return renderTemplate(c, "sign-in", nil)
// }
//
// func createAccountHandler(c echo.Context) error {
// 	return renderTemplate(c, "create-account", nil)
// }
//
// func createQuestionHandler(c echo.Context) error {
// 	return renderTemplate(c, "create-question", nil)
// }
//
// func main() {
// 	e := initEcho()
//
// 	e.Use(middleware.Logger())
//
// 	// e.Use(log.Print(e.Request().Header.Get("HX-Request")))
//
// 	// Serve static files
// 	e.Static("/static", "ui/static")
// 	log.Print("full templates: ", fullPages)
// 	log.Print("partial templates: ", partialPages)
//
// 	// Routes
// 	e.GET("/", indexHandler)
// 	e.GET("/sign-in", signInHandler)
// 	e.GET("/create-account", createAccountHandler)
// 	e.GET("/create-question", createQuestionHandler)
//
// 	// Start server
// 	log.Println("Server is running at http://localhost:8080")
// 	e.Logger.Fatal(e.Start(":8080"))
// }

// package main
//
// import (
// 	"html/template"
// 	"log"
// 	"net/http"
// 	"os"
//
// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// )
//
// var (
// 	fullPages    map[string]*template.Template
// 	partialPages map[string]*template.Template
// )
//
// // Initialize templates
// func initTemplates() {
// 	fullPages = make(map[string]*template.Template)
// 	partialPages = make(map[string]*template.Template)
//
// 	// Determine the working directory
// 	dir, err := os.Getwd()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	// Load the base layout for full pages
// 	layout := dir + "/ui/templates/index.html"
//
// 	// Parse full pages with the layout
// 	fullPages["index"] = template.Must(template.ParseFiles(layout, dir+"/ui/templates/index.html"))
// 	fullPages["create-account"] = template.Must(template.ParseFiles(layout, dir+"/ui/templates/create-account.html"))
// 	fullPages["sign-in"] = template.Must(template.ParseFiles(layout, dir+"/ui/templates/sign-in.html"))
// 	fullPages["create-question"] = template.Must(template.ParseFiles(layout, dir+"/ui/templates/create-question.html"))
//
// 	// Parse partial pages without the layout
// 	partialPages["create-account"] = template.Must(template.ParseFiles(dir + "/ui/templates/create-account.html"))
// 	partialPages["sign-in"] = template.Must(template.ParseFiles(dir + "/ui/templates/sign-in.html"))
// 	partialPages["create-question"] = template.Must(template.ParseFiles(dir + "/ui/templates/create-question.html"))
// }
//
// // Render full or partial templates based on the HX-Request header
// func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}, hx bool) {
// 	var t *template.Template
// 	var ok bool
//
// 	if hx {
// 		t, ok = partialPages[tmpl]
// 	} else {
// 		t, ok = fullPages[tmpl]
// 	}
//
// 	if !ok {
// 		http.Error(w, "Template not found: "+tmpl, http.StatusInternalServerError)
// 		return
// 	}
//
// 	if hx {
// 		err := t.Execute(w, data)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 		}
// 	} else {
// 		err := t.ExecuteTemplate(w, "index.html", data)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 		}
// 	}
// }
//
// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	hx := r.Header.Get("HX-Request") != ""
// 	renderTemplate(w, "index", nil, hx)
// }
//
// func signInHandler(w http.ResponseWriter, r *http.Request) {
// 	hx := r.Header.Get("HX-Request") != ""
// 	renderTemplate(w, "sign-in.html", nil, hx)
// }
//
// func createAccountHandler(w http.ResponseWriter, r *http.Request) {
// 	hx := r.Header.Get("HX-Request") != ""
// 	renderTemplate(w, "create-account", nil, hx)
// }
//
// func createQuestionHandler(w http.ResponseWriter, r *http.Request) {
// 	hx := r.Header.Get("HX-Request") != ""
// 	renderTemplate(w, "create-question", nil, hx)
// }
//
// func main() {
// 	initTemplates()
//
// 	log.Print("full templates: ", fullPages)
// 	log.Print("partial templates: ", partialPages)
//
// 	e := echo.New()
// 	e.Renderer = t
//
// 	e.Use(middleware.Logger())
//
// 	// Serve static files
// 	e.Static("/static", "ui/static")
//
// 	// Home page route
// 	e.GET("/", func(c echo.Context) error {
// 		// Use the main index template with the home content template
// 		data := map[string]interface{}{
// 			"Content": "sign-in", // Home content template name
// 		}
// 		return c.Render(http.StatusOK, "index.html", data)
// 	})
// 	http.HandleFunc("/", indexHandler)
// 	http.HandleFunc("/sign-in", signInHandler)
// 	http.HandleFunc("/create-account", createAccountHandler)
// 	http.HandleFunc("/create-question", createQuestionHandler)
// 	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
//
// 	log.Println("Server is running at http://localhost:8080")
// 	http.ListenAndServe(":8080", nil)
// }

// var FullPages map[string]*template.Template
//
// var partialPages  map[string]*template.Template
//
// func initTemplates() {
//   FullPages = make(map[string]*template.Template)
//     partialPages = make(map[string]*template.Template)
//     // Load all HTML files for templates, including the layout (index.html)
//     dir, err := os.Getwd()
//     if err != nil {
//   log.Fatal(err)
//     }
//
//     layout := "ui/templates/index.html"
//     FullPages["index.html"] = template.Must(template.ParseFiles(layout, dir + "/ui/templates/index.html"))
//     FullPages["create-account.html"] = template.Must(template.ParseFiles(layout, dir + "/ui/templates/create-account.html"))
//     FullPages["sign-in.html"] = template.Must(template.ParseFiles(layout, dir + "/ui/templates/sign-in.html"))
//     FullPages["create-question.html"] = template.Must(template.ParseFiles(layout, dir + "/ui/templates/create-question.html"))
//     partialPages["create-account"] = template.Must(template.ParseFiles(dir + "/ui/templates/blocks/create-account.html"))
//     partialPages["sign-in"] = template.Must(template.ParseFiles(dir + "/ui/templates/blocks/sign-in.html"))
//     partialPages["create-question"] = template.Must(template.ParseFiles(dir + "/ui/templates/blocks/create-question.html"))
// }

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
//
// 	// Load all HTML files for templates, including the layout (index.html)
// 	t := &TemplateRegistry{
// 		templates: template.Must(template.ParseGlob(dir + "/ui/templates/*.html")),
// 	}
//
// 	log.Print("--------- Loaded Templates ----------")
//
// 	for _, tmpl := range t.templates.Templates() {
// 		log.Println(tmpl.Name())
// 	}
//
// 	log.Print("=====================================")
//
// 	e := echo.New()
// 	e.Renderer = t
//
// 	e.Use(middleware.Logger())
//
// 	// Serve static files
// 	e.Static("/static", "ui/static")
//
// 	// Home page route
// 	e.GET("/", func(c echo.Context) error {
// 		// Use the main index template with the home content template
// 		data := map[string]interface{}{
// 			"Content": "sign-in", // Home content template name
// 		}
// 		return c.Render(http.StatusOK, "index.html", data)
// 	})
//
// 	// Create Account route
// 	e.GET("/create-account", func(c echo.Context) error {
// 		data := map[string]interface{}{
// 			"Content": "create-account", // Create Account content template name
// 		}
// 		return c.Render(http.StatusOK, "create-account", data)
// 	})
//
// 	e.GET("/create-account-submit", func(c echo.Context) error {
// 		data := map[string]interface{}{
// 			"Content": "create-account-submit", // Create Account content template name
// 		}
// 		return c.Render(http.StatusOK, "index.html", data)
// 	})
//
// 	// Sign-In route
// 	e.GET("/sign-in", func(c echo.Context) error {
// 		data := map[string]interface{}{
// 			"Content": "sign-in", // Sign-In content template name
// 		}
// 		return c.Render(http.StatusOK, "sign-in", data)
// 	})
//
// 	// Create Question route
// 	e.GET("/create-question", func(c echo.Context) error {
// 		data := map[string]interface{}{
// 			"Content": "create-question", // Create Question content template name
// 		}
// 		return c.Render(http.StatusOK, "index.html", data)
// 	})
//
// 	e.Logger.Fatal(e.Start(":8080"))
// }

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
