package routers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func GamesRouter(r *echo.Group) {
	r.POST("/game", createGame)
	r.GET("", getGamesPage)
	r.GET("/ui/create-game", getCreateGameUI)

	r.GET("/all", getAllGames)
}

func getGamesPage(c echo.Context) error {
	return controllers.RenderTemplate(c, "games", nil)
}

type Game struct {
	gameId      sql.NullInt64
	dateCreated sql.NullTime
	gameName    sql.NullString
	creatorId   sql.NullInt64
}

func getAllGames(c echo.Context) error {
	query := `
      SELECT * FROM games;
  `

	rows, err := db.Sqlite.Query(query, nil)
	if err != nil {
		fmt.Println("error querying db")
		return err
	}
	var games []Game

	for rows.Next() {
		var (
			gameId      sql.NullInt64
			dateCreated sql.NullTime
			gameName    sql.NullString
			creatorId   sql.NullInt64
		)
		if err := rows.Scan(&gameId, &dateCreated, &gameName, &creatorId); err != nil {
			fmt.Println("error scanning rows:", err)
			return err
		}
		game := Game{
			gameId:      gameId,
			dateCreated: dateCreated,
			gameName:    gameName,
			creatorId:   creatorId,
		}
		games = append(games, game)
	}

	fmt.Println(games)

	return c.HTML(http.StatusOK, "done")
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

func getCreateGameUI(c echo.Context) error {
	return controllers.RenderTemplate(c, "create-game", nil)
}
