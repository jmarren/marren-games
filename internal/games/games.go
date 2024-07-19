package games

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmarren/marren-games/internal/db"
)

func GetAllGames() {
}

func GetUserGames(myUserId int) ([]Game, error) {
	fail := func(err error) ([]Game, error) {
		return nil, fmt.Errorf("error (GetUserGames): %v", err)
	}
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
		return fail(err)
	}

	var games []Game

	for rows.Next() {
		var gameIdRaw sql.NullInt64
		var gameNameRaw sql.NullString
		var totalMembers sql.NullInt64

		err := rows.Scan(&totalMembers, &gameIdRaw, &gameNameRaw)
		if err != nil {
			return fail(err)
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

	return games, nil
}

func DeleteGame() {
}

func CreateGame(tx *sql.Tx, userId int, gameName string) (int64, error) {
	// make sql.Named variables from userId, gameName, creator_id
	userIdArg := sql.Named("user_id", userId)
	gameNameArg := sql.Named("game_name", gameName)
	result, err := tx.Exec(`
    INSERT INTO games (creator_id, name) VALUES (:my_user_id, :my_game_name);
    `, userIdArg, gameNameArg)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	fmt.Println("result: ", result)
	gameId, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err)
		return 0, errors.New("an error occurred")
	}

	AddPlayer(tx, gameId, userId)

	return gameId, nil
}

func GetGame() {
}
