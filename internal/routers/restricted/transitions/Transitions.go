package transitions

import (
	"fmt"
	"math/rand"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/labstack/echo/v4"
)

type TemplateData struct {
	Data interface{}
}

func TransitionRouter(r *echo.Group) {
	r.GET("/:target-page", func(c echo.Context) error {
		// Add Header
		c.Response().Header().Set(echo.HeaderCacheControl, "max-age=15000")
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

		num := rangeIn(1, 4)

		fmt.Println("Random num:", num)

		if num == 1 {
			return controllers.RenderTemplate(c, "slide-out-to-right", data)
		}
		if num == 2 {
			return controllers.RenderTemplate(c, "fade-out", data)
		}
		return controllers.RenderTemplate(c, "spin-and-shrink", data)
	})
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}
