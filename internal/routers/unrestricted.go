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

// Simplified version of the Answer struct
type SimplifiedAnswer struct {
	AnswerText       string
	AnswererID       int64
	AnswererUsername string
	QuestionID       int64
}

// Function to convert Answer to SimplifiedAnswer
func simplifyAnswer(a *Answer) *SimplifiedAnswer {
	return &SimplifiedAnswer{
		AnswerText:       a.AnswerText.String,
		AnswererID:       a.AnswererID.Int64,
		AnswererUsername: a.AnswererUsername.String,
		QuestionID:       a.QuestionID.Int64,
	}
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

					// Combine main query with WithQueries
					query := GetFullQuery(routeConfig.query, []string{routeConfig.withQuery})

					// Perform Query
					results, string, err := db.QueryWithMultipleNamedParams(query, params, routeConfig.createNewSlice, routeConfig.typ)
					if err != nil {
						fmt.Println(string)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}
					fmt.Println("results in route:", results)

					// Dynamically handle the type specified in routeConfig.typ
					resultsValue := reflect.ValueOf(results)
					// var simplifiedAnswers []SimplifiedAnswer

					pageDataSlice := reflect.MakeSlice(reflect.SliceOf(routeConfig.concreteType), 0, 0)

					// Check if the result is a slice
					if resultsValue.Kind() == reflect.Slice {
						for i := 0; i < resultsValue.Len(); i++ {
							item := resultsValue.Index(i).Interface()
							fmt.Printf("Item %d: %+v\n", i, item)

							// dereference the pointer to get the underlying struct for each slice item
							dereferencedItem := reflect.Indirect(reflect.ValueOf(item)).Interface()
							fmt.Printf("Item %d: %+\n", i, dereferencedItem)

							// var oneAnswer Answer

							dereferencedItemValue := reflect.ValueOf(dereferencedItem)

							if dereferencedItemValue.Type().ConvertibleTo(routeConfig.concreteType) {
								pageData := reflect.ValueOf(dereferencedItem).Convert(routeConfig.concreteType)
								fmt.Println("pageData: ", pageData)
								pageDataSlice = reflect.Append(pageDataSlice, pageData)
								// oneAnswer = concreteType.Interface().(Answer)
								// fmt.Println("oneAnswer: ", oneAnswer)
								// simplified := simplifyAnswer(&oneAnswer)
								// simplifiedAnswers = append(simplifiedAnswers, *simplified)
								// fmt.Printf("Simplified Item %d: %+v\n", i, *simplified)
							} else {
								fmt.Println("Unexpected type")
							}

							// resultValue := reflect.ValueOf(newSlice)
							// if resultValue.Type().ConvertibleTo(typ) {
							//   concreteType := resultValue.Convert(typ)
							//   results = reflect.Append(results, concreteType)
							// } else {
							//   fmt.Println("Unexpected type")
							// }

							// Simplify the struct
							// if answer, ok := dereferencedItem.(Answer); ok {
							// 	simplified := simplifyAnswer(&answer)
							// 	simplifiedAnswers = append(simplifiedAnswers, *simplified)
							// 	fmt.Printf("Simplified Item %d: %+v\n", i, *simplified)
							// } else {
							// 	fmt.Println("Unexpected type")
							// }

						}
					} else {
						fmt.Println("Unexpected result type")
					}

					return c.String(http.StatusOK, fmt.Sprintf("Results: %+v", pageDataSlice))
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
