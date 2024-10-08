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
	Username        string
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

type TemplateName string

const (
	IndexTemplate          TemplateName = "index"
	SignInTemplate         TemplateName = "sign-in"
	CreateAccountTemplate  TemplateName = "create-account"
	ProfileTemplate        TemplateName = "profile"
	CreateQuestionTemplate TemplateName = "create-question"
	GamePlay               TemplateName = "gameplay"
)

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
		"games.html",
		"gameplay.html",
		"question.html",
		"results.html",
		"create-game.html",
		"friends.html",
		"upload-profile-photo.html",
		"create-account-err.html",
		"slide-out-to-right.html",
		"side-bar.html",
		"login-success.html",
		"search-results.html",
		"profile-photo-viewer.html",
		"other-user-profile.html",
		"add-friend-button.html",
		"remove-friend-button.html",
		"request-sent-button.html",
		"search-bar.html",
		"invite-friends.html",
		"delete-invite-button.html",
		"invite-friend-button.html",
		"profile-after-login.html",
		"index-after-logout.html",
		"spin-and-shrink.html",
		"fade-out.html",
		"no-question-yet.html",
		"top-bar.html",
		"loader.html",
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
func RenderTemplate(c echo.Context, partialTemplate string, data interface{}) error {
	currentHeadersMap := make(map[string]string)
	headers := []string{echo.HeaderCacheControl, echo.HeaderVary}
	for _, header := range headers {
		currentHeader := c.Response().Header().Get(header)
		currentHeadersMap[header] = currentHeader
	}
	defaultHeadersMap := map[string]string{
		echo.HeaderCacheControl: "no-cache, private",
		echo.HeaderVary:         "Hx-Request, X-Username",
	}

	for k, v := range currentHeadersMap {
		if v == "" {
			c.Response().Header().Set(k, defaultHeadersMap[k])
		} else {
			c.Response().Header().Set(k, currentHeadersMap[k])
		}
	}
	hx := c.Request().Header.Get("Hx-Request") == "true"

	if hx {
		err := c.Render(http.StatusOK, partialTemplate, data)
		if err != nil {
			fmt.Println("Error rendering template from hx-request:", err)
			return err
		}
		return err
	}
	// Full Page Reload: Render the base layout with the partial content
	fmt.Println("data: ", data)

	// get usename
	username, ok := auth.GetFromClaims(auth.Username, c).(string)
	if !ok {
		username = ""
	}

	pageData := PageData{
		Username:        username,
		Title:           "Marren Games - " + partialTemplate,
		PartialTemplate: partialTemplate,
		Data:            data,
	}
	fmt.Println(pageData)
	err := c.Render(http.StatusOK, "base", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func IndexHandler(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "max-age=86400, public")
	c.Response().Header().Set(echo.HeaderVary, "Hx-Request")
	return RenderTemplate(c, "index", nil)
}

func SignInHandler(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "max-age=86400, public")
	c.Response().Header().Set(echo.HeaderVary, "Hx-Request")
	return RenderTemplate(c, "sign-in", nil)
}

func CreateAccountHandler(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "max-age=86400, public")
	c.Response().Header().Set(echo.HeaderVary, "Hx-Request")
	return RenderTemplate(c, "create-account", nil)
}

func CreateQuestionHandler(c echo.Context) error {
	return RenderTemplate(c, "create-question", nil)
}

func CreateAccountSubmitHandler(c echo.Context) error {
	jwt, registrationErr := auth.RegisterUser(c.FormValue("username"), c.FormValue("password"), c.FormValue("email"))

	if registrationErr != nil {
		data := struct {
			Error string
		}{
			Error: registrationErr.Error(),
		}

		fmt.Println(data)
		err := c.HTML(http.StatusConflict, registrationErr.Error())
		if err != nil {
			fmt.Println("error rendering err message: ", err)
		}
		fmt.Println("returning: ", err)
		return err
	}

	cookie := &http.Cookie{
		Name:    "auth",
		Value:   jwt,
		Expires: time.Now().Add(24 * time.Hour),
	}

	c.SetCookie(cookie)
	username := c.FormValue("username")

	data := struct {
		Username     string
		PhotoVersion int
	}{
		Username:     username,
		PhotoVersion: 0,
	}
	return RenderTemplate(c, "upload-profile-photo", data)
}

func LoginHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	jwt, err := auth.AuthenticateUser(username, password)
	if err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred during authentication.")
	}
	cookie := &http.Cookie{
		Name:     "auth",
		Value:    jwt,
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	}

	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/auth/profile?add-top-right=true")
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
		Username:      username.(string),
		GameCompleted: true,
	}
	return RenderTemplate(c, "profile", data)
}
