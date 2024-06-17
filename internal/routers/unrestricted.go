package routers

import (
	"database/sql"
	"fmt"
	"net/http"
	_ "net/http"
	"reflect"
	"strconv"
	"strings"

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

// func QueryTestHandler(group *echo.Group) {
// 	// var roureconfigs routeConfigs
// 	routeConfigs := GetRouteConfigs()
//
// 	for _, routeConfig := range routeConfigs {
// 		if routeConfig.method == "GET" {
// 			group.GET(routeConfig.path,
// 				func(c echo.Context) error {
// 					var params []interface{}
// 					for _, param := range routeConfig.queryParams {
// 						params = append(params, c.QueryParam(param))
// 					}
// 					fmt.Println(params)
// 					var output string
// 					err := db.Sqlite.QueryRow(routeConfig.query, params...).Scan(&output)
// 					if err != nil {
// 						return c.String(http.StatusInternalServerError, "failed to execute query")
// 					}
// 					return c.String(http.StatusOK, output)
// 				})
// 		}
// 	}
// }

func QueryTestHandler(group *echo.Group) {
	// var roureconfigs routeConfigs
	routeConfigs := GetRouteConfigs()

	for _, routeConfig := range routeConfigs {
		switch routeConfig.method {
		case "GET":
			group.GET(routeConfig.path,
				func(c echo.Context) error {
					var params []interface{}
					for _, paramConfig := range routeConfig.queryParams {
						paramValue := c.QueryParam(paramConfig.Name)
						convertedValue, err := convertType(paramValue, paramConfig.Type)
						if err != nil {
							errorMessage := fmt.Errorf("**** error converting type: %s\n| parameter name: %s\n| parameter value: %s\n| paramConfig.Type: %s ", err, paramConfig.Name, paramValue, paramConfig.Type)
							fmt.Println(errorMessage)
							return c.String(http.StatusBadRequest, error.Error(errorMessage))
						}
						params = append(params, convertedValue)
					}
					fmt.Println(params)

					outputRows, err := db.Sqlite.Query(routeConfig.query, params...)
					if err != nil {
						return err
					}

					cols, err := outputRows.Columns()
					if err != nil {
						return err
					}
					colLen := len(cols)
					vals := make([]interface{}, colLen)
					valPtrs := make([]interface{}, colLen)

					for outputRows.Next() {
						for i := range cols {
							valPtrs[i] = &vals[i]
						}
						err := outputRows.Scan(valPtrs...)
						if err != nil {
							fmt.Println(err)
							return err
						}
						for i, col := range cols {
							val := vals[i]

							b, ok := val.([]byte)
							var v interface{}
							if ok {
								v = string(b)
							} else {
								v = val
							}
							fmt.Println(col, v)
						}
					}

					return c.String(http.StatusOK, strings.Join([]string{" ", " "}, "\n"))
				})
		case "POST":
			group.POST(routeConfig.path,
				func(c echo.Context) error {
					var params []interface{}
					for _, paramConfig := range routeConfig.queryParams {
						paramValue := c.QueryParam(paramConfig.Name)
						convertedValue, err := convertType(paramValue, paramConfig.Type)
						if err != nil {
							errorMessage := fmt.Errorf("**** error converting type: %s\n| parameter name: %s\n| parameter value: %s\n| paramConfig.Type: %s ", err, paramConfig.Name, paramValue, paramConfig.Type)
							fmt.Println(errorMessage)
							return c.String(http.StatusBadRequest, error.Error(errorMessage))
						}
						params = append(params, convertedValue)
					}
					fmt.Println(params)

					var result sql.Result
					result, err := db.Sqlite.Exec(routeConfig.query, params...)
					if err != nil {
						log.Error(err)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}

					fmt.Println(result)

					return c.String(http.StatusOK, "Record created successfully")

					// for outputRows.Next() {
					// 	err := outputRows.Scan(&output)
					// 	if err != nil {
					// 		log.Error(err)
					// 		return c.String(http.StatusInternalServerError, "failed to execute query")
					// 	}
					// }
					//
					// return c.String(http.StatusOK, output)
				})
		}
	}
}

// Utility function to convert query parameter from string to specified type
func convertType(value string, targetType reflect.Kind) (interface{}, error) {
	switch targetType {
	case reflect.Int:
		return strconv.Atoi(value)
	case reflect.String:
		return value, nil
	// Add more type cases as needed (e.g., float64, bool, etc.)
	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType)
	}
}
