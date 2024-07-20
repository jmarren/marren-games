package games

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func InvitesRouter(r *echo.Group) {
	// TODO
	// r.GET("", getGameInvitesByUserId)
	r.DELETE("", declineGameInvite)
	r.POST("/:user-id", invitePlayerToGame)
	r.DELETE("/:user-id", deleteGameInvite)
}

// TODO
// func getGameInvitesByUserId(c echo.Context) error {
// }

func invitePlayerToGame(c echo.Context) error {
	newPlayerId := c.Param("user-id")
	gameId := c.Param("game-id")
	newPlayerIdInt, err := strconv.Atoi(newPlayerId)
	if err != nil {
		return fmt.Errorf("user id query param not convertible to int: %v ", err)
	}
	gameIdInt, err := strconv.Atoi(gameId)
	if err != nil {
		return fmt.Errorf("game-id param not convertible to int: %v ", err)
	}

	newPlayerArg := sql.Named("new_player_id", newPlayerId)
	gameIdArg := sql.Named("game_id", gameIdInt)

	query := `
      INSERT INTO user_game_invites (user_id, game_id)
  VALUES(:new_player_id, :game_id);
  `
	_, err = db.Sqlite.Exec(query, newPlayerArg, gameIdArg)
	if err != nil {
		fmt.Println("error adding user to user_game_invites")
		return err
	}

	data := struct {
		GameId int
		UserId int
	}{
		GameId: gameIdInt,
		UserId: newPlayerIdInt,
	}
	return controllers.RenderTemplate(c, "delete-invite-button", data)
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
		playerId, err = strconv.Atoi(c.Param("user-id"))
		if err != nil {
			fmt.Println("playerId not convertible to int")
			return err
		}
	}

	gameId := c.Param("game-id")
	gameIdInt, err := strconv.Atoi(gameId)
	if err != nil {
		return fmt.Errorf("game-id param not convertible to int %v", err)
	}

	// playerIdInt, err := strconv.Atoi(playerId)
	// if err != nil {
	// 	fmt.Errorf("game-id param not convertible to int")
	// }
	playerIdArg := sql.Named("player_id", playerId)
	gameIdArg := sql.Named("game_id", gameId)

	query := `
      DELETE FROM user_game_invites
      WHERE user_id = :player_id AND game_id = :game_id;
  `
	_, err = db.Sqlite.Exec(query, playerIdArg, gameIdArg)
	if err != nil {
		fmt.Println("error removing user from user_game_invites")
		return err
	}

	if shortenedUrl == "/auth/games" {
		return c.HTML(http.StatusOK, `declined`)
	}

	// playerIdStr := strconv.Itoa(playerId)

	data := struct {
		GameId int
		UserId int
	}{
		GameId: gameIdInt,
		UserId: playerId,
	}
	return controllers.RenderTemplate(c, "invite-friend-button", data)
}

func declineGameInvite(c echo.Context) error {
	// get necessary data (userId and )
	gameId, err := strconv.Atoi(c.Param("game-id"))
	if err != nil {
		return fmt.Errorf("gameId: not convertible to int: %v ", err)
	}
	userId := auth.GetFromClaims(auth.UserId, c)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))
	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("players router transaction: %v", err)
	}

	// convert to sql.NamedArg
	gameIdArg := sql.Named("game_id", gameId)
	userIdArg := sql.Named("user_id", userId)

	// delete user from invites
	query := `
    DELETE FROM user_game_invites
    WHERE user_id = :user_id
      AND game_id = :game_id;
  `
	// perform query
	_, err = tx.ExecContext(ctx, query, gameIdArg, userIdArg)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("playersRouter, acceptInvite(): error deleting from invites: %v", err)
	}
	tx.Commit()

	return c.HTML(http.StatusOK, "invite declined")
}
