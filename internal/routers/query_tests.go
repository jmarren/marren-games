package routers

type routeConfig struct {
	path        string
	method      string
	query       string
	queryParams []string
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
				path:        "/get-username-with-id",
				method:      "GET",
				query:       `SELECT username FROM users WHERE id = ?`,
				queryParams: []string{"id"},
			},
		})
	return routeConfigs
}
