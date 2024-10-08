package routers

import (
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/labstack/echo/v4"
)

func UnrestrictedRoutes(group *echo.Group) {
	group.GET("/", controllers.IndexHandler)
	group.GET("/sign-in", controllers.SignInHandler)
	group.GET("/create-account", controllers.CreateAccountHandler)
	group.POST("/login", controllers.LoginHandler)
	group.POST("/create-account-submit", controllers.CreateAccountSubmitHandler)
}

type NamedParam struct {
	Name  string
	Value interface{}
}

type TemplateData struct {
	Data interface{}
}
