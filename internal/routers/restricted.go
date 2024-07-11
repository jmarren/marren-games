package routers

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
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
		SigningKey:  []byte("secret"), // TODO
		TokenLookup: "cookie:auth",
	}
	r.Use(echojwt.WithConfig(jwtConfig))

	transitionGroup := r.Group("/transition")
	TransitionRouter(transitionGroup)
	gamesGroup := r.Group("/games")
	GamesRouter(gamesGroup)
	profileGroup := r.Group("/profile")
	ProfileRouter(profileGroup)
	friendsGroup := r.Group("/friends")
	FriendsRouter(friendsGroup)

	r.GET("/create-question", controllers.CreateQuestionHandler)

	RestrictedRouteConfigs := GetRestrictedRouteConfigs()

	for _, routeConfig := range RestrictedRouteConfigs {
		switch routeConfig.method {
		case GET:
			r.GET(routeConfig.path,

				func(c echo.Context) error {
					// convert params to the type specified in config
					data, err := GetRequestWithDbQuery(routeConfig, c)
					if err != nil {
						fmt.Println("error performing dynamic query: ", err)
						return c.String(http.StatusInternalServerError, "error")
					}
					// Create a TemplateData struct to pass to the template
					templateData := TemplateData{
						Data: data,
					}
					return controllers.RenderTemplate(c, routeConfig.partialTemplate, templateData)
				})
		}
	}
}
