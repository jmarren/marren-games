package games

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func InvitesRouter(r *echo.Group) {
	r.POST("", invitePlayerToGame)
	r.DELETE("", deleteGameInvite)
}

func invitePlayerToGame(c echo.Context) error {
	newPlayerId := c.QueryParam("user-id")
	gameId := c.QueryParam("game-id")

	newPlayerArg := sql.Named("new_player_id", newPlayerId)
	gameIdArg := sql.Named("game_id", gameId)

	query := `
      INSERT INTO user_game_invites (user_id, game_id)
  VALUES(:new_player_id, :game_id);
  `
	_, err := db.Sqlite.Exec(query, newPlayerArg, gameIdArg)
	if err != nil {
		fmt.Println("error adding user to user_game_invites")
		return err
	}

	return c.HTML(http.StatusOK, `<button hx-delete="/auth/games/game/invites?user-id=`+newPlayerId+`&game-id=`+gameId+`" hx-swap="outerHTML">
       Remove
      </button>`)
}

func deleteGameInvite(c echo.Context) error {
	fromUrl := c.Request().Header.Get("Hx-Current-Url")
	fmt.Println("----------- fromUrl: ", fromUrl)
	shortenedUrl := fromUrl[len(fromUrl)-11:]

	fmt.Println("----------- shortenedUrl: ", fromUrl)

	var playerId int

	if shortenedUrl == "/auth/games" {
		playerId = auth.GetFromClaims(auth.UserId, c).(int)
	} else {
		var err error
		playerId, err = strconv.Atoi(c.QueryParam("user-id"))
		if err != nil {
			fmt.Println("playerId not convertible to int")
			return err
		}
	}

	gameId := c.QueryParam("game-id")

	playerIdArg := sql.Named("player_id", playerId)
	gameIdArg := sql.Named("game_id", gameId)

	query := `
      DELETE FROM user_game_invites
      WHERE user_id = :player_id AND game_id = :game_id;
  `
	_, err := db.Sqlite.Exec(query, playerIdArg, gameIdArg)
	if err != nil {
		fmt.Println("error removing user from user_game_invites")
		return err
	}

	if shortenedUrl == "/auth/games" {
		return c.HTML(http.StatusOK, `declined`)
	}

	playerIdStr := strconv.Itoa(playerId)

	return c.HTML(http.StatusOK, `<button hx-post="/auth/games/game/invites?user-id=`+playerIdStr+`&game-id=`+gameId+`" hx-swap="outerHTML">
       + Invite
      </button>`)
}
