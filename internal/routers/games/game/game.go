package games

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

func GameRouter(r *echo.Group) {
	r.POST("", createGame)
	// r.GET("", getGame)

	invitesGroup := r.Group("/invites")
	InvitesRouter(invitesGroup)

	questionsGroup := r.Group("/questions")
	QuestionsRouter(questionsGroup)
}

func createGame(c echo.Context) error {
	userId := auth.GetFromClaims("UserId", c)
	gameName := c.FormValue("game-name")

	myUserIdArg := sql.Named("my_user_id", userId)
	gameNameArg := sql.Named("my_game_name", gameName)

	fmt.Println("userId: ", userId)
	fmt.Println("gameName: ", gameName)

	if gameName == "" {
		fmt.Println(" NAME NOT PROVIDED")
		return c.HTML(http.StatusBadRequest, "please provide a name")
	}
	result, err := db.Sqlite.Exec(`
    INSERT INTO games (creator_id, name) VALUES (:my_user_id, :my_game_name);
    `, myUserIdArg, gameNameArg)
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

	query := `
    INSERT INTO user_game_membership (user_id, game_id)
    VALUES (:my_user_id, :game_id);
  `
	gameIdArg := sql.Named("game_id", gameId)

	_, err = db.Sqlite.Exec(query, myUserIdArg, gameIdArg)
	if err != nil {
		fmt.Println("error adding creator to user_game_membership")
		return errors.New("internal server error")
	}

	data := struct {
		GameId int64
	}{
		GameId: gameId,
	}

	return controllers.RenderTemplate(c, "invite-friends", data)
}
