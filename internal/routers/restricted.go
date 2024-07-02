package routers

import (
	"database/sql"
	"fmt"
	"net/http"
	_ "net/http"
	"reflect"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func (c ClaimArgConfig) getValue(context echo.Context) interface{} {
	return auth.GetFromClaims(c.claim, context)
}

func (u UrlParamArgConfig) getValue(context echo.Context) string {
	return context.QueryParam(string(u.urlParam))
}

func RestrictedRoutes(r *echo.Group) {
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtCustomClaims)
		},
		SigningKey:  []byte("secret"),
		TokenLookup: "cookie:auth",
	}

	r.Use(echojwt.WithConfig(jwtConfig))

	r.GET("/test", func(c echo.Context) error {
		return c.String(200, "You are authenticated")
	})

	// r.GET("/profile", controllers.ProfileHandler)
	r.GET("/create-question", controllers.CreateQuestionHandler)

	RestrictedRouteConfigs := GetRestrictedRouteConfigs()

	for _, routeConfig := range RestrictedRouteConfigs {
		switch routeConfig.method {
		case GET:
			r.GET(routeConfig.path,

				func(c echo.Context) error {
					// convert params to the type specified in config
					params, err := GetParamsFromUrlAndClaims(routeConfig.claimArgConfigs, routeConfig.urlParamArgConfigs, c)
					if err != nil {
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println("params: ", params)

					// Combine main query with WithQueries
					query := GetFullQuery(routeConfig.query, routeConfig.withQueries)
					fmt.Printf("\nquery: %v", query)

					// Perform Query
					results, string, err := db.DynamicQuery(query, params, routeConfig.typ)
					// results, string, err := db.QueryWithMultipleNamedParams(query, params, routeConfig.typ)
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

					fmt.Printf("resultsValue: %v", resultsValue)

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
								fmt.Printf("\ndereferencedItemValue.Type(): %v", dereferencedItemValue.Type())
								fmt.Println("Unexpected type")
							}
						}
					} else {
						fmt.Println("Unexpected result type")
					}

					simplifiedFields := GetSimplifiedFields(reflect.TypeOf(routeConfig.typ))

					simplified := SimplifySqlResult(results, simplifiedFields)

					// Create a TemplateData struct to pass to the template
					templateData := TemplateData{
						Data: simplified,
					}
					return controllers.RenderTemplate(c, routeConfig.partialTemplate, templateData)
				})
		}
	}
}

func GetFullQuery(mainQuery string, withQueries []string) string {
	query := ""
	for _, withQuery := range withQueries {
		query += "\n" + db.WithQueries.GetWithQuery(withQuery)
	}
	return mainQuery + "\n" + query
}

func GetParamsFromUrlAndClaims(claimArgConfigs []ClaimArgConfig, urlParamConfigs []UrlParamArgConfig, c echo.Context) ([]sql.NamedArg, error) {
	// Get params from urlParamArgConfigs and claimArgConfigs
	var params []sql.NamedArg

	// convert urlParamArgConfigs into their specified type, convert to namedParams and append to params
	for _, urlParamConfig := range urlParamConfigs {
		value := urlParamConfig.getValue(c)
		convertedValue, err := ConvertType(string(value), urlParamConfig.Type)
		if err != nil {
			fmt.Println("error converting urlparms to specified Type", err)
			return params, err
		}
		namedParam := sql.Named(string(urlParamConfig.urlParam), convertedValue)
		params = append(params, namedParam)
	}

	// convert claimArgConfigs into their specified type, convert to namedParams and append to params
	for _, claimConfig := range claimArgConfigs {
		value := claimConfig.getValue(c)
		convertedValue, err := ConvertType(value, claimConfig.Type)
		if err != nil {
			fmt.Println("error converting claims to specified Type", err)
			return params, err
		}
		namedParam := sql.Named(string(claimConfig.claim), convertedValue)
		params = append(params, namedParam)
	}

	return params, nil
}
