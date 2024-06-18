package routers

import (
	"database/sql"
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
					var params []interface{}

					for _, paramConfig := range routeConfig.queryParams {
						paramValue := c.QueryParam(paramConfig.Name)
						convertedValue, err := convertType(paramValue, paramConfig.Type)
						if err != nil {
							errorMessage := fmt.Errorf("**** error converting type: %s\n| parameter name: %s\n| parameter value: %s\n| paramConfig.Type: %s ", err, paramConfig.Name, paramValue, paramConfig.Type)
							fmt.Println(errorMessage)
							return c.String(http.StatusBadRequest, error.Error(errorMessage))
						}
						// namedParam := NamedParam{Name: paramConfig.Name, Value: convertedValue}
						// namedParams = append(namedParams, namedParam)
						namedParam := sql.Named(paramConfig.Name, convertedValue)
						params = append(params, namedParam)
					}
					fmt.Println(params)

					// Perform Query
					outputRows, err := db.Sqlite.Query(routeConfig.query, params...)
					if err != nil {
						fmt.Println(err)
						return err
					}

					cols, err := outputRows.Columns()
					if err != nil {
						fmt.Println(err)
						return err
					}
					colLen := len(cols)
					vals := make([]interface{}, colLen)
					valPtrs := make([]interface{}, colLen)

					response := ""

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
							response = fmt.Sprintf("%v\n %v: %v", response, col, v)
						}
						response = fmt.Sprintf("%v\n-----------", response)
					}

					return c.String(http.StatusOK, response)
				})

			// POST Requests
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
						// namedParam := fmt.Sprintf("$%s=%v", paramConfig.Name, convertedValue)
						namedParam := sql.Named(paramConfig.Name, convertedValue)
						params = append(params, namedParam)
					}
					fmt.Println(routeConfig.query)
					fmt.Println(params)
					var result sql.Result
					result, err := db.Sqlite.Exec(routeConfig.query, params...)
					if err != nil {
						log.Error(err)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}

					fmt.Println(result)

					return c.String(http.StatusOK, "Record created successfully")
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
