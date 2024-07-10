package routers

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func GamesRouter(r *echo.Group) {
	r.POST("/create-game", createGame)

	routeConfigs := GetGamesRoutes()

	for _, routeConfig := range routeConfigs {
		switch routeConfig.method {
		case GET:
			r.GET(routeConfig.path,
				func(c echo.Context) error {
					fmt.Println(" hit new gamesRouter")

					if routeConfig.query == "" {
						return controllers.RenderTemplate(c, routeConfig.partialTemplate, nil)
					}

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

func GetGamesRoutes() RouteConfigs {
	return CreateNewRouteConfigs(
		[]RouteConfig{
			{
				path:                    "",
				method:                  GET,
				claimArgConfigs:         []ClaimArgConfig{},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				withQueries:             []string{},
				query: `SELECT (
                 json_group_array(
                    json_object(
                      id, name, creator_id
                    )
                 ) 
                ) as Games 
                 FROM games;`,
				typ: struct {
					Games []struct {
						GameId    int    `json:"id"`
						GameName  string `json:"name"`
						CreatorId int    `json:"creator_id"`
					}
				}{},
				partialTemplate: "games",
			},
			{
				path:   "/create-game",
				method: GET,
				claimArgConfigs: []ClaimArgConfig{
					{claim: auth.UserId, Type: reflect.Int},
				},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				withQueries:             []string{},
				query:                   "",
				typ:                     struct{}{},
				partialTemplate:         "create-game",
			},
		})
}

func createGame(c echo.Context) error {
	userId := auth.GetFromClaims("UserId", c)
	gameName := c.FormValue("game-name")
	friends := c.FormValue("friends")

	fmt.Println("userId: ", userId)
	fmt.Println("gameName: ", gameName)
	fmt.Println("friends: ", friends)

	if gameName == "" {
		return c.HTML(http.StatusBadRequest, "please provide a name")
	}
	result, err := db.Sqlite.Exec(`
    INSERT INTO games (creator_id, name) VALUES (?, ?);
    `, userId, gameName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("result: ", result)
	gameId, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err)
		return errors.New("an error occurred")
	}

	data := struct {
		GameId int64
	}{
		GameId: gameId,
	}

	return controllers.RenderTemplate(c, "create-question", data)
}
