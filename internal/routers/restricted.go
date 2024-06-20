package routers

import (
	"database/sql"
	"fmt"
	"net/http"
	_ "net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func (c ClaimArgConfig) getValue(context echo.Context) string {
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

	r.GET("/profile", controllers.ProfileHandler)
	r.GET("/create-question", controllers.CreateQuestionHandler)

	RestrictedRouteConfigs := GetRestrictedRouteConfigs()

	for _, config := range RestrictedRouteConfigs {
		switch config.method {
		case GET:
			r.GET(config.path,
				func(c echo.Context) error {
					params, err := GetParamsFromUrlAndClaims(config.query.claimArgConfigs, config.query.urlParamArgConfigs, c)
					if err != nil {
						fmt.Println(err)
						return c.String(http.StatusBadRequest, "error getting params")
					}
					fmt.Println(params)

					query := GetFullQuery(config.query.mainQuery, config.query.withQueries)

					fmt.Println(query)

					queryResult, err := db.QueryWithMultipleNamedParams(query, params)
					if err != nil {
						return c.String(http.StatusBadRequest, "error querying database")
					}

					fmt.Println(queryResult)

					return c.String(http.StatusOK, "still building")
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
