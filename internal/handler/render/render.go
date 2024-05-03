package render

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	t := &Templates{
		templates: template.Must(template.ParseGlob("ui/templates/*.html")),
	}
	t.templates = template.Must(t.templates.ParseGlob("ui/templates/blocks/*.html"))
	return t
}

type PageData struct {
	Options []string
}
