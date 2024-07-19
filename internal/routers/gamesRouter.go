package routers

import (
	"context"
	"fmt"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/jmarren/marren-games/internal/games"
	"github.com/labstack/echo/v4"
)

func GamesRouter(r *echo.Group) {
	r.POST("/game", createGameHandler)
	// r.POST("/game/invites", invitePlayerToGame)
	// r.DELETE("/game/invites", deleteGameInvite)
	r.GET("/game/play/:game-id", getPlayPage)
	r.GET("", getGamesPage)
	// r.GET("/ui/create-game", getCreateGameUI)
	r.POST("/game/questions", createQuestion)
	r.POST("/game/:game-id/question/:question-id/answers", createAnswer)
	// r.GET("/ui/invite-friends", getInviteFriendUI)
	r.POST("/game/players", acceptGameInvite)
	r.GET("/all", getAllGames)
	r.GET("/game/ui/create-question", getCreateQuestionUI)
}

func createGameHandler(c echo.Context) error {
	// helper to return errors
	fail := func(err error) error {
		return fmt.Errorf("createGameHandler: %v", err)
	}

	// Get context values
	userId := auth.GetFromClaims(auth.UserId, c).(int)
	gameName := c.FormValue("game-name")

	// create context
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	defer cancel()

	// begin transaction
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		fail(err)
	}
	defer tx.Rollback()

	// create game
	gameId, err := games.CreateGame(tx, userId, gameName)
	if err != nil {
		fail(err)
	}

	// construct data for template
	data := struct {
		GameId int64
	}{
		GameId: gameId,
	}

	return controllers.RenderTemplate(c, "invite-friends", data)
}
