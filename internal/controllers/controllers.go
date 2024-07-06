package controllers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/awssdk"
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
		err := c.Render(http.StatusOK, "base", pageData)
		if err != nil {
			fmt.Println(err)
		}
		return err
	}
}

func IndexHandler(c echo.Context) error {
	return RenderTemplate(c, "index", nil)
}

func SignInHandler(c echo.Context) error {
	return RenderTemplate(c, "sign-in", nil)
}

func CreateAccountHandler(c echo.Context) error {
	return RenderTemplate(c, "create-account", nil)
}

func CreateQuestionHandler(c echo.Context) error {
	return RenderTemplate(c, "create-question", nil)
}

func CreateAccountSubmitHandler(c echo.Context) error {
	// err := c.Request().ParseMultipartForm(1000)
	// fmt.Println(c.Request().Header)
	// if err != nil {
	// 	fmt.Println("error parsing multipartform ", err)
	// }

	registrationError := auth.RegisterUser(c.FormValue("username"), c.FormValue("password"), c.FormValue("email"))

	fmt.Println("hit create account submit handler")

	// profilePhoto, err := c.FormFile("profileImage")
	// if err != nil {
	// 	fmt.Println("error getting FormFile: ", err)
	// 	return err
	// }
	//
	// src, err := profilePhoto.Open()
	// if err != nil {
	// 	fmt.Println("error opening file: ", err)
	// 	return err
	// }
	// defer src.Close()
	//
	// // Destination
	// dst, err := os.Create(profilePhoto.Filename)
	// if err != nil {
	// 	fmt.Println("error creating file with os.Create(): ", err)
	// 	return err
	// }
	// defer dst.Close()
	//
	// // Copy
	// if _, err = io.Copy(dst, src); err != nil {
	// 	fmt.Println("error copying file: ", err)
	// 	return err
	// }
	//
	if registrationError == nil {
		return RenderTemplate(c, "upload-profile-photo", CreateAccountSuccessData{
			Username: c.FormValue("username"),
		})
	}

	data := CreateAccountData{
		Username: c.FormValue("username"),
		Email:    c.FormValue("email"),
		Error:    registrationError,
	}

	return RenderTemplate(c, "create-account", data)
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
		Username:      username.(string),
		GameCompleted: true,
	}

	id, err := db.GetUserIdFromUsername(username.(string))
	if err != nil {
		return err
	}

	fmt.Println("\n----------- UserId: ", id, "-------------")

	fmt.Println("\n----------- Username: ", username, "---------------")
	fmt.Println("")

	return RenderTemplate(c, "profile", data)
}

func UploadProfilePhotoHandler(c echo.Context) error {
	fmt.Println(c.Request().Header)

	f, err := c.FormFile("profileImage")
	if err != nil {
		fmt.Println("error during c.FormFile: ", err)
		return err
	}

	fmt.Println("\n\nfile: ", f)

	fmt.Println(reflect.TypeOf(f))

	uploadErr := awssdk.UploadToS3(f)
	if uploadErr != nil {
		fmt.Println("uploadError uploading to s3: ", uploadErr)
		return uploadErr
	}

	// for _, file := range files {
	// 	// Source
	// 	src, err := file.Open()
	// 	if err != nil {
	// 		fmt.Println("error during file.Open(): ", err)
	// 		return err
	// 	}
	// 	defer src.Close()
	//
	// 	// Destination
	// 	dst, err := os.Create(file.Filename)
	// 	fmt.Println("error during os.Create(): ", err)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer dst.Close()
	//
	// 	// Copy
	// 	if _, err = io.Copy(dst, src); err != nil {
	// 		fmt.Println("error while copying: ", err)
	// 		return err
	// 	}
	//
	// }
	return c.Redirect(http.StatusFound, "/auth/profile")
}
