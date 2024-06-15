package routers

import (
	_ "net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func RestrictedRoutes(r *echo.Group) {
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtCustomClaims)
		},
		SigningKey:  []byte("secret"),
		TokenLookup: "cookie:auth",
	}

	r.Use(echojwt.WithConfig(jwtConfig))

	r.GET("/index", controllers.IndexHandler)
	r.GET("/test", func(c echo.Context) error {
		return c.String(200, "You are authenticated")
	})

	r.GET("/profile", controllers.ProfileHandler)
}
