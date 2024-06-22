package routers

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

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
	queryTest := group.Group("/query")
	QueryTestHandler(queryTest)
}

type NamedParam struct {
	Name  string
	Value interface{}
}

func QueryTestHandler(group *echo.Group) {
	routeConfigs := GetRouteConfigs()

	for _, routeConfig := range routeConfigs {
		switch routeConfig.method {
		// GET Requests
		case "GET":
			group.GET(routeConfig.path,

				func(c echo.Context) error {
					// convert params to the type specified in config
					params, err := GetParamsFromUrlAndClaims(routeConfig.claimArgConfigs, routeConfig.urlParamArgConfigs, c)
					if err != nil {
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println("params: ", params)

					query := GetFullQuery(routeConfig.query, []string{routeConfig.withQuery})

					results, string, err := db.QueryWithMultipleNamedParams(query, params, routeConfig.createNewSlice, routeConfig.typ)
					if err != nil {
						fmt.Println(string)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}

					fmt.Println("results in route:", results)

					// Type assertion for usage
					if concreteResults, ok := results.([]*Answer); ok {
						for _, answer := range concreteResults {
							fmt.Printf("Answer: %+v\n", *answer)
						}
					} else {
						fmt.Println(reflect.TypeOf(results))
						fmt.Println("Type assertion failed")
					}

					return c.String(http.StatusOK, " ")
				})

			// POST Requests
		case "POST":
			group.POST(routeConfig.path,
				func(c echo.Context) error {
					params, err := GetParamsFromUrlAndClaims(routeConfig.claimArgConfigs, routeConfig.urlParamArgConfigs, c)
					if err != nil {
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println("params: ", params)

					query := GetFullQuery(routeConfig.query, []string{routeConfig.withQuery})

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

// Utility function to convert query parameter from string to specified type
func ConvertType(value string, targetType reflect.Kind) (interface{}, error) {
	switch targetType {
	case reflect.Int:
		result, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println("error converting string to int", err)
			fmt.Println("** Trying to convert ", value, " to int")
			return nil, err
		}
		return result, nil
	case reflect.String:
		return value, nil
	// Add more type cases as needed (e.g., float64, bool, etc.)
	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType)
	}
}
