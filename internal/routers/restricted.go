package routers

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	_ "net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
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
					// Get params from urlParamArgConfigs and claimArgConfigs
					var params []interface{}

					// convert urlParamArgConfigs into their specified type, convert to namedParams and append to params
					for _, urlParamConfig := range config.query.urlParamArgConfigs {
						value := urlParamConfig.getValue(c)
						convertedValue, err := ConvertType(string(value), urlParamConfig.Type)
						if err != nil {
							return err
						}
						namedParam := sql.Named(string(urlParamConfig.urlParam), convertedValue)
						params = append(params, namedParam)
					}

					// convert claimArgConfigs into their specified type, convert to namedParams and append to params
					for _, claimConfig := range config.query.claimArgConfigs {
						value := claimConfig.getValue(c)
						convertedValue, err := ConvertType(value, claimConfig.Type)
						if err != nil {
							return err
						}
						namedParam := sql.Named(string(claimConfig.claim), convertedValue)
						params = append(params, namedParam)
					}

					fmt.Println(params)
					return c.String(http.StatusOK, "still building")
				})
		}
	}
}

type withQueries struct {
  queries *map[string]string
}

func GetWithQuery(fileName string) withQueries {
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



}


func *




//
// func GetQueryArgs(q queryParam, c echo.Context) interface{} {
// 	switch q.Source {
// 	case FromClaims:
// 		return auth.GetFromClaims(q.Name, c)
// 	case FromUrlParams:
// 		return c.QueryParam(q.Name)
// 	default:
// 		return []interface{}{}
// 	}
// }
