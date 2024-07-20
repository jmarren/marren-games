package games

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
)

func PlayersRouter(r *echo.Group) {
	// TODO
	// r.GET("", getPlayersByGameId)
	// r.DELETE(":user-id", deletePlayerFromGame)
	// r.DELETE("", declineInvite)
	r.POST("", acceptInvite)
}

func acceptInvite(c echo.Context) error {
	// get userId and gameId
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

	query = `
    INSERT INTO user_game_membership (user_id, game_id)
    VALUES (:user_id, :game_id);
  `

	// perform query
	_, err = tx.ExecContext(ctx, query, gameIdArg, userIdArg)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("playersRouter, acceptInvite(): error adding user to game: %v", err)
	}

	tx.Commit()
	return c.HTML(http.StatusOK, "invite accepted!")
}

//
// func declineInvite(c echo.Context) error {
//   // get necessary data (userId and )
//   gameId, err := strconv.Atoi(c.Param("game-id"))
// 	if err != nil {
// 		return fmt.Errorf("gameId: not convertible to int: %v ", err)
// 	}
// 	userId := auth.GetFromClaims(auth.UserId, c)
//
// 	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))
// 	defer cancel()
//
// 	tx, err := db.Sqlite.BeginTx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("players router transaction: %v", err)
// 	}
//
// 	// convert to sql.NamedArg
// 	gameIdArg := sql.Named("game_id", gameId)
// 	userIdArg := sql.Named("user_id", userId)
//
// 	// delete user from invites
// 	query := `
//     DELETE FROM user_game_invites
//     WHERE user_id = :user_id
//       AND game_id = :game_id;
//   `
// 	// perform query
// 	_, err = tx.ExecContext(ctx, query, gameIdArg, userIdArg)
// 	if err != nil {
// 		tx.Rollback()
// 		return fmt.Errorf("playersRouter, acceptInvite(): error deleting from invites: %v", err)
// 	}
//
//
//
//
//
// }
//

// TODO
// func getPlayersByGameId(c echo.Context) error {
// }
//
// func deletePlayerFromGame(c echo.Context) error {}
//
// func addPlayerToGame(c echo.Context) error {}
