package routers

import "reflect"

type routeConfig struct {
	path        string
	method      string
	query       string
	queryParams []ParamConfig
}

type ParamConfig struct {
	Name string
	Type reflect.Kind
}

type routeConfigs []*routeConfig

func CreateNewRouteConfigs(r []routeConfig) routeConfigs {
	var routeConfigs routeConfigs
	for _, routeConfig := range r {
		routeConfigs = append(routeConfigs, &routeConfig)
	}
	return routeConfigs
}

func CreateNewRouteConfig() *routeConfig {
	return &routeConfig{}
}

func GetRouteConfigs() routeConfigs {
	routeConfigs := CreateNewRouteConfigs(
		[]routeConfig{
			{
				path:   "/get-username-with-id",
				method: "GET",
				query:  `SELECT username FROM users WHERE id = ?`,
				queryParams: []ParamConfig{
					{Name: "id", Type: reflect.Int},
				},
			},
			{
				path:   "/questions-by-user-id",
				method: "GET",
				query: `SELECT question_text, date_created 
                FROM questions
                INNER JOIN users
                  ON users.id = questions.asker_id
                WHERE questions.asker_id = ?;`,
				queryParams: []ParamConfig{
					{Name: "id", Type: reflect.Int},
				},
			},
			{
				path:   "/create-question",
				method: "GET",
				query: `INSERT INTO questions (asker_id, question_text)
                VALUES (?,?);`,
				queryParams: []ParamConfig{
					{Name: "id", Type: reflect.Int},
					{Name: "text", Type: reflect.String},
				},
			},
		})
	return routeConfigs
}
