package games

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func InvitePlayer(c echo.Context, newPlayerId int, gameId int) error {
	// newPlayerId := c.QueryParam("user-id")
	// gameId := c.QueryParam("game-id")

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

	data := struct {
		NewPlayerId int
		GameId      int
	}{
		NewPlayerId: newPlayerId,
		GameId:      gameId,
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

func GetGameInvites(c echo.Context, userId int) ([]GameInvite, error) {
	myUserIdArg := sql.Named("my_user_id", userId)

	// // Get All Invites for the User
	query := `
	       SELECT name, creator_id, game_id, users.username AS creator_name
         FROM user_game_invites
         LEFT JOIN games
          ON games.id = game_id
         LEFT JOIN users
          ON creator_id = users.id
         WHERE user_id = :my_user_id;
	 `

	var gameInvites []GameInvite

	rows, err := db.Sqlite.Query(query, myUserIdArg)
	if err != nil {
		fmt.Println("error querying for game invites: ", err)
		return nil, err
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
			return nil, err
		}
		gameInvites = append(gameInvites, GameInvite{
			GameName:    gameNameRaw.String,
			CreatorId:   creatorIdRaw.Int64,
			GameId:      gameIdRaw.Int64,
			CreatorName: creatorNameRaw.String,
		})

	}
	return gameInvites, nil
}
