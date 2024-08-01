package restricted

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/routers/friends"
	"github.com/jmarren/marren-games/internal/routers/profile"
	"github.com/jmarren/marren-games/internal/routers/restricted/games"
	"github.com/jmarren/marren-games/internal/routers/restricted/transitions"
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
	// Use jwt with claims for authentication
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtCustomClaims)
		},
		SigningKey:  []byte(os.Getenv("JWTSECRET")), // TODO
		TokenLookup: "cookie:auth",
	}
	r.Use(echojwt.WithConfig(jwtConfig))

	// Middleware that adds custom username and Vary Headers
	// that tell the browser to look for changes in X-Username and Hx-Request
	// Headers before using cached content
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			username, ok := auth.GetFromClaims(auth.Username, c).(string)
			if !ok {
				return errors.New("username from claims not assertible to string")
			}

			c.Response().Header().Set("X-Username", username)
			c.Response().Header().Set(echo.HeaderVary, "X-Username, Hx-Request")
			return next(c)
		}
	})

	transitionGroup := r.Group("/transition")
	transitions.TransitionRouter(transitionGroup)
	gamesGroup := r.Group("/games")
	games.GamesRouter(gamesGroup)
	profileGroup := r.Group("/profile")
	profile.ProfileRouter(profileGroup)
	friendsGroup := r.Group("/friends")
	friends.FriendsRouter(friendsGroup)

	r.GET("/create-question", controllers.CreateQuestionHandler)
}
