package render

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates *template.Template
}

type AnswerQuestionData struct {
	Question string
	Choices  []string
}

type CreateAccountData struct {
	Username string
	Email    string
}

// func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
// 	return t.templates.ExecuteTemplate(w, "index.html", data)
// }

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Render the specific block based on the route
	switch c.Request().URL.Path {
	case "/":
		return t.templates.ExecuteTemplate(w, "index.html", data)
	case "/sign-in":
		return t.templates.ExecuteTemplate(w, "sign-in.html", data)
	case "/create-account":
		return t.templates.ExecuteTemplate(w, "create-account.html", data)
	default:
		return t.templates.ExecuteTemplate(w, "index.html", data)
	}
}

// func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
// 	// Parse the specific block template
// 	block, err := template.ParseFiles(fmt.Sprintf("ui/templates/blocks/%s.html", name))
// 	if err != nil {
// 		return err
// 	}
//
// 	// Add the block template to the parsed templates under the name "content"
// 	tmpl := template.Must(t.templates.Clone())
// 	tmpl, err = tmpl.AddParseTree("content", block.Tree)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Execute the index template
// 	return tmpl.ExecuteTemplate(w, "index.html", data)
// }

func NewTemplates() *Templates {
	t := &Templates{
		templates: template.Must(template.ParseGlob("ui/templates/blocks/*.html")),
	}
	t.templates = template.Must(t.templates.ParseGlob("ui/templates/*.html"))
	return t
}

type PageData struct {
	Options []string
}
