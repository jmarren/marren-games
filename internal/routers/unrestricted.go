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
					params, err := GetParamsFromUrlAndClaims(routeConfig.claimArgConfigs, routeConfig.urlParamArgConfigs, c)
					if err != nil {
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println("params: ", params)

					// Combine main query with WithQueries
					query := GetFullQuery(routeConfig.query, routeConfig.withQueries)

					// Perform Query
					results, string, err := db.QueryWithMultipleNamedParams(query, params, routeConfig.typ)
					if err != nil {
						fmt.Println(string)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}
					fmt.Println("results in route:", results)

					// Dynamically handle the type specified in routeConfig.typ
					resultsValue := reflect.ValueOf(results)

					dataType := reflect.TypeOf(routeConfig.typ)
					sliceOfDataType := reflect.SliceOf(dataType)
					concreteDataSlice := reflect.MakeSlice(sliceOfDataType, 0, 0)

					// Check if the result is a slice
					// If it is, iterate through the slice and convert the items to the concrete type specified in routeConfig
					if resultsValue.Kind() == reflect.Slice {
						for i := 0; i < resultsValue.Len(); i++ {
							item := resultsValue.Index(i).Interface()

							// dereference the pointer to get the underlying struct for each slice item
							dereferencedItem := reflect.Indirect(reflect.ValueOf(item)).Interface()

							// Convert the dereferencedItem to the concrete type specified in routeConfig
							dereferencedItemValue := reflect.ValueOf(dereferencedItem)

							if dereferencedItemValue.Type().ConvertibleTo(dataType) {
								concrete := reflect.ValueOf(dereferencedItem).Convert(dataType)
								concreteDataSlice = reflect.Append(concreteDataSlice, concrete)

							} else {
								fmt.Println("Unexpected type")
							}
						}
					} else {
						fmt.Println("Unexpected result type")
					}

					// Simplify the results from sql generics to primitive types
					simplifiedFields := GetSimplifiedFields(dataType)
					simplifiedStructType := reflect.StructOf(simplifiedFields)
					sliceOfSimplifiedStructType := reflect.SliceOf(simplifiedStructType)
					simplifiedResults := reflect.MakeSlice(sliceOfSimplifiedStructType, 0, 0)

					for i := 0; i < concreteDataSlice.Len(); i++ {
						simplifiedResult := SimplifySqlResult(concreteDataSlice.Index(i).Interface())
						simplifiedResults = reflect.Append(simplifiedResults, reflect.ValueOf(simplifiedResult))
					}
					fmt.Println("simplifiedResults: ", simplifiedResults)

					// Create a TemplateData struct to pass to the template
					templateData := TemplateData{
						Data: concreteDataSlice.Interface(),
					}
					return controllers.RenderTemplate(c, routeConfig.partialTemplate, templateData)
				})

			// POST Requests
		case POST:
			group.POST(routeConfig.path,
				func(c echo.Context) error {
					params, err := GetParamsFromUrlAndClaims(routeConfig.claimArgConfigs, routeConfig.urlParamArgConfigs, c)
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
