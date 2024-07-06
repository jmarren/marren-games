package routers

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func UnrestrictedRoutes(group *echo.Group) {
	group.GET("/", controllers.IndexHandler)
	group.GET("/sign-in", controllers.SignInHandler)
	group.GET("/create-account", controllers.CreateAccountHandler)
	group.POST("/login", controllers.LoginHandler)
	group.POST("/create-account-submit", controllers.CreateAccountSubmitHandler)
	unrestricted := group.Group("/unrestricted")
	QueryTestHandler(unrestricted)
}

type NamedParam struct {
	Name  string
	Value interface{}
}

type TemplateData struct {
	Data interface{}
}

func QueryTestHandler(group *echo.Group) {
	routeConfigs := GetRouteConfigs()

	for _, routeConfig := range routeConfigs {
		switch routeConfig.method {
		// GET Requests
		case GET:
			group.GET(routeConfig.path,

				func(c echo.Context) error {
					// convert params to the type specified in config
					params, err := GetParams(routeConfig.claimArgConfigs, routeConfig.urlQueryParamArgConfigs, routeConfig.urlPathParamArgConfigs, c)
					if err != nil {
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println("params: ", params)

					// Combine main query with WithQueries
					query := GetFullQuery(routeConfig.query, routeConfig.withQueries)

					fmt.Println("query: ", query, "\nrouteConfig.typ:", routeConfig.typ)

					// Perform Query
					results, string, err := db.DynamicQuery(query, params, routeConfig.typ)
					if err != nil {
						fmt.Println(string)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}

					simplifiedFields := GetSimplifiedFields(reflect.TypeOf(routeConfig.typ))

					simplified := SimplifySqlResult(results, simplifiedFields)

					// Create a TemplateData struct to pass to the template
					templateData := TemplateData{
						Data: simplified,
					}
					return controllers.RenderTemplate(c, routeConfig.partialTemplate, templateData)
				})

			// POST Requests
		case POST:
			group.POST(routeConfig.path,
				func(c echo.Context) error {
					params, err := GetParams(routeConfig.claimArgConfigs, routeConfig.urlQueryParamArgConfigs, routeConfig.urlPathParamArgConfigs, c)
					if err != nil {
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println("params: ", params)

					query := GetFullQuery(routeConfig.query, routeConfig.withQueries)

					fmt.Println(params)
					result, err := db.ExecTestWithNamedParams(query, params)
					if err != nil {
						log.Error(err)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}

					fmt.Println(string(result))

					return c.String(http.StatusOK, "Record created successfully")
				})

		}
	}
}
