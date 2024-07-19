package games

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func GetPlayPageContext(tx *sql.Tx, userId int, gameId int) (bool, bool, error) {
	// helper function for errors
	fail := func(err error) (bool, bool, error) {
		return false, false, fmt.Errorf("GetPlayPageContext %v", err)
	}

	// convert vars to sql.NamedArg for query
	userIdArg := sql.Named("user_id", userId)
	gameIdArg := sql.Named("game_id", gameId)

	// Query
	query := `
    SELECT (
      CASE WHEN
    (
    SELECT user_id
    FROM current_askers
    WHERE current_askers.game_id = :game_id) = :user_id THEN 1
    ELSE 0
    END
    ) AS is_asker,
    (
      CASE WHEN (
        SELECT COUNT(*)
        FROM questions
        WHERE game_id = :game_id
          AND DATE(date_created) = DATE('now')
        ) > 0 THEN 1
      ELSE 0
      END)
    AS todays_question_created
    FROM current_askers;
  `

	// Variables for scanning results
	var isAsker sql.NullInt64
	var todaysQuestionCreated sql.NullInt64

	// create context
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()
	// perform query
	row := tx.QueryRowContext(ctx, query, gameIdArg, userIdArg)

	// scan into vars
	err := row.Scan(&isAsker, &todaysQuestionCreated)
	if err != nil {
		fail(err)
	}

	// convert isAsker and todaysQuestionCreated to booleans
	var isAskerBool bool
	var todaysQuestionCreatedBool bool

	if isAsker.Int64 == 0 {
		isAskerBool = false
	} else {
		isAskerBool = true
	}
	if todaysQuestionCreated.Int64 == 0 {
		todaysQuestionCreatedBool = false
	} else {
		todaysQuestionCreatedBool = true
	}

	return isAskerBool, todaysQuestionCreatedBool, nil
}
