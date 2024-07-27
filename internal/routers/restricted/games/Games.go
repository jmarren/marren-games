package games

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func GamesRouter(r *echo.Group) {
	r.GET("", getGames)
	r.POST("", createGame)
	r.GET("/create", getCreateGameUI)

	gameGroup := r.Group("/:game-id")
	GameRouter(gameGroup)
}

func getGames(c echo.Context) error {
	// Get User Id
	myUserId := auth.GetFromClaims(auth.UserId, c)

	// Get All Games where User is a Member
	// create named arg
	myUserIdArg := sql.Named("my_user_id", myUserId)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	// begin the Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}

	defer tx.Rollback()
	type Game struct {
		GameId           int64
		GameName         string
		GameTotalMembers int64
	}

	var games []Game

	query := `
SELECT user_game_membership.game_id, games.name, member_counts.total_members 
  FROM user_game_membership
JOIN games
 	ON games.id = user_game_membership.game_id
JOIN (
      SELECT game_id , COUNT(user_id) AS total_members
  FROM user_game_membership
  GROUP BY game_id) AS member_counts
    ON member_counts.game_id = user_game_membership.game_id
  WHERE user_game_membership.user_id = :my_user_id;
  `

	rows, err := tx.QueryContext(ctx, query, myUserIdArg)
	if err != nil {
		fmt.Println("error querying db for user's game ids:", err)
		return err
	}

	for rows.Next() {
		var gameIdRaw sql.NullInt64
		var gameNameRaw sql.NullString
		var totalMembers sql.NullInt64

		err := rows.Scan(&gameIdRaw, &gameNameRaw, &totalMembers)
		if err != nil {
			fmt.Println("error scanning game_id into gameId variable: ", err)
			return err
		}
		if !gameIdRaw.Valid || !gameNameRaw.Valid {
			fmt.Println("gameId not valid:", gameIdRaw, gameNameRaw)
		}

		games = append(games, Game{
			GameId:           gameIdRaw.Int64,
			GameName:         gameNameRaw.String,
			GameTotalMembers: totalMembers.Int64,
		})
	}

	fmt.Println("%%%%%%%%%% games: ", games, "%%%%%%%%%%%%")

	// // Get All Invites for the User
	query = `
	       SELECT name, creator_id, game_id, users.username AS creator_name
         FROM user_game_invites
         LEFT JOIN games
          ON games.id = game_id
         LEFT JOIN users
          ON creator_id = users.id
         WHERE user_id = :my_user_id;
	 `

	var gameInvites []GameInvite

	rows, err = tx.QueryContext(ctx, query, myUserIdArg)
	if err != nil {
		tx.Rollback()
		fmt.Println("error querying for game invites: ", err)
		return err
	}

	for rows.Next() {
		var (
			gameNameRaw    sql.NullString
			creatorIdRaw   sql.NullInt64
			gameIdRaw      sql.NullInt64
			creatorNameRaw sql.NullString
		)

		err := rows.Scan(&gameNameRaw, &creatorIdRaw, &gameIdRaw, &creatorNameRaw)
		if err != nil {
			fmt.Println("error scanning invites into vars: ", err)
			return err
		}
		gameInvites = append(gameInvites, GameInvite{
			GameName:    gameNameRaw.String,
			CreatorId:   creatorIdRaw.Int64,
			GameId:      gameIdRaw.Int64,
			CreatorName: creatorNameRaw.String,
		})

	}
	fmt.Println("%%%%%%%%%% Game Invites: ", gameInvites, "  %%%%%%%%%%%%%")

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("games, getGames(), error commiting tx: %v", err)
	}
	data := struct {
		CurrentGames []Game
		GameInvites  []GameInvite
	}{
		CurrentGames: games,
		GameInvites:  gameInvites,
	}
	return controllers.RenderTemplate(c, "games", data)
}

func getCreateGameUI(c echo.Context) error {
	return controllers.RenderTemplate(c, "create-game", nil)
}

func createGame(c echo.Context) error {
	userId := auth.GetFromClaims("UserId", c)
	gameName := c.FormValue("game-name")

	myUserIdArg := sql.Named("my_user_id", userId)
	gameNameArg := sql.Named("my_game_name", gameName)

	if gameName == "" {
		fmt.Println(" NAME NOT PROVIDED")
		return c.HTML(http.StatusBadRequest, "please provide a name")
	}

	// create context for query
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	// begin the Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}

	defer tx.Rollback()

	// Insert new game into games table
	result, err := tx.ExecContext(ctx, `
    INSERT INTO games (creator_id, name) VALUES (:my_user_id, :my_game_name);
    `, myUserIdArg, gameNameArg)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error: Games.go: inserting into games, %v", err)
	}
	gameId, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return errors.New("an error occurred")
	}

	data := struct {
		GameId int64
	}{
		GameId: gameId,
	}

	tx.Commit()
	c.Response().Header().Set("Hx-Push-Url", fmt.Sprintf("/auth/games/%v/invite-friends", gameId))
	return controllers.RenderTemplate(c, "invite-friends", data)
}
