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

	query := `
  WITH game_ids AS (
    SELECT game_id
    FROM user_game_membership
    WHERE user_id = :my_user_id
  ),
  game_names_and_ids AS (
    SELECT name, game_id
    FROM games
    WHERE id = (SELECT game_id FROM game_ids)
  )
   SELECT COUNT(user_id) AS members, game_id, name
   FROM user_game_membership
   LEFT JOIN games
    ON games.id = game_id
   WHERE game_id = (SELECT game_id FROM game_names_and_ids)
   GROUP BY game_id;
  `
	rows, err := db.Sqlite.Query(query, myUserIdArg)
	if err != nil {
		fmt.Println("error querying db for user's game ids:", err)
		return err
	}

	type Game struct {
		GameId           int64
		GameName         string
		GameTotalMembers int64
	}

	var games []Game

	for rows.Next() {
		var gameIdRaw sql.NullInt64
		var gameNameRaw sql.NullString
		var totalMembers sql.NullInt64

		err := rows.Scan(&totalMembers, &gameIdRaw, &gameNameRaw)
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

	rows, err = db.Sqlite.Query(query, myUserIdArg)
	if err != nil {
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

	c.Response().Header().Set("Hx-Push-Url", fmt.Sprintf("/auth/games/%v/invite-friends", gameId))

	data := struct {
		GameId int64
	}{
		GameId: gameId,
	}

	return controllers.RenderTemplate(c, "invite-friends", data)
}
