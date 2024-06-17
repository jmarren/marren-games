package routers

import (
	"fmt"
	"net/http"
	_ "net/http"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func UnrestrictedRoutes(group *echo.Group) {
	group.GET("/", controllers.IndexHandler)
	group.GET("/sign-in", controllers.SignInHandler)
	group.GET("/create-account", controllers.CreateAccountHandler)
	group.POST("/login", controllers.LoginHandler)
	group.POST("/create-account-submit", controllers.CreateAccountSubmitHandler)
	queryTest := group.Group("/query")
	QueryTestHandler(queryTest)
}

func QueryTestHandler(group *echo.Group) {
	// var roureconfigs routeConfigs
	routeConfigs := GetRouteConfigs()

	for _, routeConfig := range routeConfigs {
		if routeConfig.method == "GET" {
			group.GET(routeConfig.path,
				func(c echo.Context) error {
					var params []interface{}
					for _, param := range routeConfig.queryParams {
						params = append(params, c.QueryParam(param))
					}
					fmt.Println(params)
					response := db.QueryRowHandler(routeConfig.query, params...)
					return c.String(http.StatusOK, response)
				})
		}
	}
}
