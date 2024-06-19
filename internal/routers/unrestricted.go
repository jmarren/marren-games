package routers

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"os"
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

type NamedParam struct {
	Name  string
	Value interface{}
}

func QueryTestHandler(group *echo.Group) {
	routeConfigs := GetRouteConfigs()

	// get current working directory
	cwd, _ := os.Getwd()
	fmt.Println("current directory: ", cwd)

	sqlDir := os.DirFS(cwd + "/internal/sql")
	fmt.Println("sqlDir: ", sqlDir)

	sqlFiles, err := fs.ReadDir(sqlDir, ".")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(sqlFiles)

	withQueries := make(map[string]string)

	for _, file := range sqlFiles {
		query := ""
		reader, err := fs.ReadFile(sqlDir, file.Name())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("reader: ", reader)
		bytesReader := bytes.NewReader(reader)
		scanner := bufio.NewScanner(bytesReader)

		for scanner.Scan() {
			query += scanner.Text()
		}
		queryName := strings.Trim(file.Name(), ".sql")
		withQueries[queryName] = query
	}

	fmt.Println("withQueries: ", withQueries)

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
						namedParam := sql.Named(paramConfig.Name, convertedValue)
						params = append(params, namedParam)
					}
					fmt.Println(params)

					withQuery := withQueries[routeConfig.WithQuery]

					query := withQuery + " " + routeConfig.query

					fmt.Println(query)

					// Perform Query
					outputRows, err := db.Sqlite.Query(query, params...)
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
						convertedValue, err := ConvertType(paramValue, paramConfig.Type)
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
func ConvertType(value string, targetType reflect.Kind) (interface{}, error) {
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
