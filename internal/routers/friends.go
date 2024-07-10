package routers

import (
	"fmt"
	"net/http"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/labstack/echo/v4"
)

func FriendsRouter(r *echo.Group) {
	routeConfigs := GetFriendsRoutes()

	for _, routeConfig := range routeConfigs {
		switch routeConfig.method {
		case GET:
			r.GET(routeConfig.path,

				func(c echo.Context) error {
					// convert params to the type specified in config
					data, err := GetRequestWithDbQuery(routeConfig, c)
					if err != nil {
						fmt.Println("error performing dynamic query: ", err)
						return c.String(http.StatusInternalServerError, "error")
					}
					// Create a TemplateData struct to pass to the template
					templateData := TemplateData{
						Data: data,
					}
					return controllers.RenderTemplate(c, routeConfig.partialTemplate, templateData)
				})
		}
	}

	for _, routeConfig := range routeConfigs {
		switch routeConfig.method {
		case GET:

			r.GET(routeConfig.path,
				func(c echo.Context) error {
					fmt.Println(" hit new gamesRouter")

					if routeConfig.query == "" {
						return controllers.RenderTemplate(c, routeConfig.partialTemplate, nil)
					}

					data, err := GetRequestWithDbQuery(routeConfig, c)
					if err != nil {
						fmt.Println("error performing dynamic query: ", err)
						return c.String(http.StatusInternalServerError, "error")
					}

					// Create a TemplateData struct to pass to the template
					templateData := TemplateData{
						Data: data,
					}

					return controllers.RenderTemplate(c, routeConfig.partialTemplate, templateData)
				})
		}
	}
}

func GetFriendsRoutes() RouteConfigs {
	return CreateNewRouteConfigs(
		[]RouteConfig{
			{
				path:                    "",
				method:                  GET,
				claimArgConfigs:         []ClaimArgConfig{},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				withQueries:             []string{},
				query:                   ``,
				typ:                     struct{}{},
				partialTemplate:         "friends",
			},
		})
}
