package routers

import (
	"reflect"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
)

type PageData struct {
	title    string
	template controllers.TemplateName
	data     DataInterface
}

type DataInterface interface {
	executeQuery(string)
}

type RouteMethod string

const (
	GET    RouteMethod = "GET"
	POST   RouteMethod = "POST"
	PUT    RouteMethod = "PUT"
	DELETE RouteMethod = "DELETE"
)

type UrlParam string

const (
	user_id  UrlParam = "user_id"
	username UrlParam = "username"
)

type QueryParam interface {
	getValue()
}

type ClaimArgConfig struct {
	claim auth.ClaimsType
	Type  reflect.Kind
}

type UrlParamArgConfig struct {
	urlParam UrlParam
	Type     reflect.Kind
}

type Vote struct {
	voterId       int
	voterUsername string
}

func GetRestrictedRouteConfigs() []*RouteConfig {
	return CreateNewRouteConfigs(
		[]RouteConfig{
			{
				path:   "/profile",
				method: GET,
				claimArgConfigs: []ClaimArgConfig{
					{claim: auth.Username, Type: reflect.String},
				},
				urlParamArgConfigs: []UrlParamArgConfig{},
				withQueries:        []string{},
				query: `SELECT username, email,
                  CASE
                    WHEN questions.asker_id = users.id  THEN 1
                    ELSE 0
                  END AS is_asker
                FROM users
                LEFT JOIN questions
                  ON users.id
                WHERE users.username = :Username;`,
				partialTemplate: "profile",
				typ: struct {
					Username string
					Email    string
					IsAsker  int
				}{},
			},
		},
	)
}
