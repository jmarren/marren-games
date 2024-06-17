package controllers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

// TemplateRegistry is a custom HTML template renderer for Echo framework
type TemplateRegistry struct {
	templates *template.Template
}

// Render implements the Echo Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var TmplRegistry *TemplateRegistry

type PageData struct {
	Title           string
	PartialTemplate string
	Data            interface{}
}

type CreateAccountData struct {
	Username string
	Email    string
	Error    error
}
type SignInData struct {
	Username string
}
type CreateAccountSuccessData struct {
	Username string
}

type ProfileData struct {
	Username      string
	GameCompleted bool
}

// Initialize templates
func InitTemplates() {
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
		"profile.html",
		"user-profile.html",
	}

	// Create a base layout template
	templates := template.New("base").Funcs(template.FuncMap{})

	// Parse base layout
	templates = template.Must(templates.ParseFiles(dir + "/" + basePath + "base.html"))

	// Parse all partial templates (blocks)
	for _, partial := range partials {
		fmt.Println(partial)
		templates = template.Must(templates.ParseFiles(dir + "/" + basePath + partial))
	}

	// Set up the global template registry
	TmplRegistry = &TemplateRegistry{templates: templates}
}

// Render full or partial templates based on the HX-Request header
func renderTemplate(c echo.Context, partialTemplate string, data interface{}) error {
	hx := c.Request().Header.Get("HX-Request") != ""

	if hx {
		// HTMX Request: Render only the partial content
		return c.Render(http.StatusOK, partialTemplate, data)
	} else {
		// Full Page Reload: Render the base layout with the partial content
		fmt.Println("data: ", data)
		pageData := PageData{
			Title:           "Marren Games",
			PartialTemplate: partialTemplate,
			Data:            data,
		}
		fmt.Println(pageData)
		return c.Render(http.StatusOK, "base", pageData)
	}
}

func IndexHandler(c echo.Context) error {
	return renderTemplate(c, "index", nil)
}

func SignInHandler(c echo.Context) error {
	return renderTemplate(c, "sign-in", nil)
}

func CreateAccountHandler(c echo.Context) error {
	return renderTemplate(c, "create-account", nil)
}

func CreateQuestionHandler(c echo.Context) error {
	return renderTemplate(c, "create-question", nil)
}

func CreateAccountSubmitHandler(c echo.Context) error {
	registrationError := auth.RegisterUser(c.FormValue("username"), c.FormValue("password"), c.FormValue("email"))

	if registrationError == nil {
		return renderTemplate(c, "create-account-success", CreateAccountSuccessData{
			Username: c.FormValue("username"),
		})
	}
	data := CreateAccountData{
		Username: c.FormValue("username"),
		Email:    c.FormValue("email"),
		Error:    registrationError,
	}

	return renderTemplate(c, "create-account", data)
}

func LoginHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	jwt, err := auth.AuthenticateUser(username, password)
	if err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred during authentication.")
	}
	cookie := &http.Cookie{
		Name:    "auth",
		Value:   jwt,
		Expires: time.Now().Add(24 * time.Hour),
	}

	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/auth/profile")
}

func LogoutHandler(c echo.Context) error {
	cookie := &http.Cookie{
		Name:    "auth",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	}
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/")
}

func ProfileHandler(c echo.Context) error {
	username := auth.GetFromClaims(auth.Username, c)
	data := ProfileData{
		Username:      username,
		GameCompleted: true,
	}

	id, err := db.GetUserIdFromUsername(username)
	if err != nil {
		return err
	}

	fmt.Println("\n----------- UserId: ", id, "-------------")

	fmt.Println("\n----------- Username: ", username, "---------------")
	fmt.Println("")

	return renderTemplate(c, "profile", data)
}

// func CreateQuestion(c echo.Context) error {
// }
