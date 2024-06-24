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

type Query struct {
	mainQuery          string
	withQueries        []string
	urlParamArgConfigs []UrlParamArgConfig
	claimArgConfigs    []ClaimArgConfig
}

// type routeConfig struct {
// 	path               string
// 	method             string
// 	withQuery          string
// 	query              string
// 	claimArgConfigs    []ClaimArgConfig
// 	urlParamArgConfigs []UrlParamArgConfig
// 	partialTemplate    string
// 	typ                interface{}
// }

// type RestrictedRouteConfig struct {
// 	path     string
// 	method   RouteMethod
// 	claims   []auth.ClaimsType
// 	pageData *PageData
// 	query    *Query
// 	typ      interface{}
// }

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
				path:               "/profile",
				method:             GET,
				claimArgConfigs:    []ClaimArgConfig{},
				urlParamArgConfigs: []UrlParamArgConfig{},
				withQueries:        []string{},
				query: `SELECT username, email, date_created
                FROM users
                WHERE id = :user_id`,
				typ: struct {
					Username    string
					Email       string
					DateCreated string
				}{},
			},
		},
	)
}
