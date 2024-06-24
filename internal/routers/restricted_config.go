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

type RestrictedRouteConfig struct {
	path     string
	method   RouteMethod
	claims   []auth.ClaimsType
	pageData *PageData
	query    *Query
	typ      interface{}
}

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

func CreateNewRestrictedRouteConfigs(r []RestrictedRouteConfig) []*RestrictedRouteConfig {
	var configs []*RestrictedRouteConfig
	for _, config := range r {
		configs = append(configs, &config)
	}
	return configs
}

type Vote struct {
	voterId       int
	voterUsername string
}

// //// Route Specific Data Structures //////
// type Answer struct {
// 	answerText       string
// 	answererId       int
// 	answererUsername string
// 	votes            []Vote
// }
//
// type Question struct {
// 	questionText  string
// 	askerId       int
// 	askerUsername string
// }
//
// type Game struct {
// 	question Question
// 	answers  []Answer
// }
// //
// type ProfileData struct {
// 	username   string
// 	todaysGame Game
// }
//
// func (p ProfileData) executeQuery(query string) {
// }

func GetRestrictedRouteConfigs() []*RestrictedRouteConfig {
	return CreateNewRestrictedRouteConfigs(
		[]RestrictedRouteConfig{
			{
				path:   "/profile",
				method: GET,
				claims: []auth.ClaimsType{auth.Username},
				pageData: &PageData{
					title:    "Profile",
					template: controllers.ProfileTemplate,
					// data: ProfileData{
					// 	username: "",
					// 	todaysGame: Game{
					// 		question: Question{
					// 			questionText:  "",
					// 			askerId:       0,
					// 			askerUsername: "",
					// 		},
					// 		// answers: []Answer{
					// 		// 	{
					// 		// 		answerText:       "",
					// 		// 		answererId:       0,
					// 		// 		answererUsername: "",
					// 		// 		votes: []Vote{
					// 		// 			{
					// 		// 				voterId:       0,
					// 		// 				voterUsername: "",
					// 		// 			},
					// 		// 		},
					// 		// 	},
					// 		// },
					// 	},
					// },
				},
				query: &Query{
					withQueries: []string{""},
					mainQuery:   "SELECT * FROM users WHERE username = :Username",
					claimArgConfigs: []ClaimArgConfig{
						{claim: auth.Username, Type: reflect.String},
					},
				},
			},
		},
	)
}
