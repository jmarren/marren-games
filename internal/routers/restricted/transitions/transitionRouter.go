package transitions

import (
	"fmt"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/labstack/echo/v4"
)

type TemplateData struct {
	Data interface{}
}

func TransitionRouter(r *echo.Group) {
	r.GET("/:target-page", func(c echo.Context) error {
		fmt.Printf("\n\nhit\n\n")
		targetPage := c.Param("target-page")
		dataStruct := struct {
			TargetPage string
		}{
			TargetPage: targetPage,
		}

		data := TemplateData{
			Data: dataStruct,
		}
		return controllers.RenderTemplate(c, "slide-out-to-right", data)
	})
}
