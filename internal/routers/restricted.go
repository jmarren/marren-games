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

func (u UrlQueryParamArgConfig) getValue(context echo.Context) string {
	return context.QueryParam(string(u.Name))
}

func (u UrlPathParamArgConfig) getValue(context echo.Context) string {
	return context.Param(string(u.Name))
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

	r.POST("/upload-profile-photo", controllers.UploadProfilePhotoHandler)
	r.GET("/create-question", controllers.CreateQuestionHandler)

	RestrictedRouteConfigs := GetRestrictedRouteConfigs()

	for _, routeConfig := range RestrictedRouteConfigs {
		switch routeConfig.method {
		case GET:
			r.GET(routeConfig.path,

				func(c echo.Context) error {
					// convert params to the type specified in config
					params, err := GetParams(routeConfig.claimArgConfigs, routeConfig.urlQueryParamArgConfigs, routeConfig.urlPathParamArgConfigs, c)
					if err != nil {
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println("params: ", params)

					// Combine main query with WithQueries
					query := GetFullQuery(routeConfig.query, routeConfig.withQueries)
					fmt.Printf("\nquery: %v", query)

					// Perform Query
					results, string, err := db.DynamicQuery(query, params, routeConfig.typ)
					if err != nil {
						fmt.Println(string)
						return c.String(http.StatusInternalServerError, "failed to execute query")
					}
					fmt.Println()
					fmt.Println("results in route:", results)

					// Dynamically handle the type specified in routeConfig.typ
					resultsValue := reflect.ValueOf(results)

					fmt.Printf("resultsValue: %v", resultsValue)

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

func GetParams(claimArgConfigs []ClaimArgConfig, urlQueryParamArgConfigs []UrlQueryParamArgConfig, urlPathParamArgConfigs []UrlPathParamArgConfig, c echo.Context) ([]sql.NamedArg, error) {
	// Get params from urlParamArgConfigs and claimArgConfigs
	var params []sql.NamedArg

	// convert urlParamArgConfigs into their specified type, convert to namedParams and append to params
	for _, urlQueryParamConfig := range urlQueryParamArgConfigs {
		value := urlQueryParamConfig.getValue(c)
		convertedValue, err := ConvertType(string(value), urlQueryParamConfig.Type)
		if err != nil {
			fmt.Println("error converting urlparms to specified Type", err)
			return params, err
		}
		namedParam := sql.Named(string(urlQueryParamConfig.Name), convertedValue)
		params = append(params, namedParam)
	}

	// claims are already typed
	for _, claimConfig := range claimArgConfigs {
		value := claimConfig.getValue(c)
		namedParam := sql.Named(string(claimConfig.claim), value)
		params = append(params, namedParam)
	}

	for _, urlPathParamArgConfig := range urlPathParamArgConfigs {
		value := urlPathParamArgConfig.getValue(c)
		convertedValue, err := ConvertType(string(value), urlPathParamArgConfig.Type)
		if err != nil {
			fmt.Println("error converting urlparms to specified Type", err)
			return params, err
		}
		namedParam := sql.Named(string(urlPathParamArgConfig.Name), convertedValue)
		params = append(params, namedParam)
	}

	return params, nil
}
