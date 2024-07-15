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
	r.POST("/game/invites", invitePlayerToGame)
	r.DELETE("/game/invites", deleteGameInvite)
	r.GET("", getGamesPage)
	r.GET("/ui/create-game", getCreateGameUI)

	r.GET("/all", getAllGames)
}

func getGamesPage(c echo.Context) error {
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

	data := struct {
		Data []Game
	}{
		Data: games,
	}

	// // Get All Invites for the User
	//  query := `
	//        SELECT * FROM
	//  `
	//

	return controllers.RenderTemplate(c, "games", data)
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

	data := struct {
		GameId int64
	}{
		GameId: gameId,
	}

	return controllers.RenderTemplate(c, "invite-friends", data)
}

func getCreateGameUI(c echo.Context) error {
	return controllers.RenderTemplate(c, "create-game", nil)
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
	playerId := c.QueryParam("user-id")
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

	return c.HTML(http.StatusOK, `<button hx-post="/auth/games/game/invites?user-id=`+playerId+`&game-id=`+gameId+`" hx-swap="outerHTML">
       + Invite
      </button>`)
}
