package routers

import (
	"database/sql"
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

type UrlPathParamArgConfig struct {
	Name string
	Type reflect.Kind
}

type UrlQueryParamArgConfig struct {
	Name UrlParam
	Type reflect.Kind
}

type Vote struct {
	voterId       int
	voterUsername string
}

func GetRestrictedRouteConfigs() []*RouteConfig {
	return CreateNewRouteConfigs(
		[]RouteConfig{
			{
				path:                    "/game/:id",
				method:                  GET,
				claimArgConfigs:         []ClaimArgConfig{},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{
					{Name: "id", Type: reflect.Int},
				},
				query: `SELECT Username
                          FROM users
                          WHERE id = :id`,
				typ: struct {
					UserId sql.NullString
				}{},
				partialTemplate: "gameplay",
			},
			{
				path:   "/create-game",
				method: GET,
				claimArgConfigs: []ClaimArgConfig{
					{claim: auth.UserId, Type: reflect.Int},
				},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				query: `SELECT Username
                      FROM users
                      WHERE id = :UserId`,
				typ: struct {
					UserId sql.NullString
				}{},
				partialTemplate: "create-game",
			},
			{
				path:   "/create-question",
				method: GET,
				claimArgConfigs: []ClaimArgConfig{
					{claim: auth.UserId, Type: reflect.Int},
				},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				query: `SELECT Username
                FROM users
                WHERE id = :UserId`,
				typ: struct {
					UserId sql.NullString
				}{},
				partialTemplate: "create-question",
			},
		},
	)
}
