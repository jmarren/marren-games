package games

import (
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/labstack/echo/v4"
)

func UiRouter(r *echo.Group) {
	r.GET("/create-game", getCreateGameUI)
}

func getCreateGameUI(c echo.Context) error {
	return controllers.RenderTemplate(c, "create-game", nil)
}
