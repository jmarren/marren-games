package routers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func GamesRouter(r *echo.Group) {
	r.POST("/game", createGame)
	r.POST("/game/invites", invitePlayerToGame)
	r.DELETE("/game/invites", deleteGameInvite)
	r.GET("/game/play/:game-id", getPlayPage)

	r.GET("", getGamesPage)
	r.GET("/ui/create-game", getCreateGameUI)
	r.POST("/game/questions", createQuestion)
	r.POST("/game/:game-id/question/:question-id/answers", createAnswer)
	// r.GET("/ui/invite-friends", getInviteFriendUI)
	r.POST("/game/players", acceptGameInvite)
	r.GET("/all", getAllGames)
	r.GET("/game/ui/create-question", getCreateQuestionUI)
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

	type GameInvite struct {
		GameName    string
		CreatorId   int64
		GameId      int64
		CreatorName string
	}

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

func acceptGameInvite(c echo.Context) error {
	newPlayerId := auth.GetFromClaims(auth.UserId, c)
	gameId := c.QueryParam("game-id")

	newPlayerIdArg := sql.Named("new_player_id", newPlayerId)
	gameIdArg := sql.Named("game_id", gameId)

	// create a context
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}

	defer tx.Rollback()

	query := `
      DELETE FROM user_game_invites
      WHERE user_id = :new_player_id AND game_id = :game_id;
  `
	_, err = db.Sqlite.ExecContext(ctx, query, newPlayerIdArg, gameIdArg)
	if err != nil {
		fmt.Println("error while deleting game invite: ", err)
		return err
	}

	// delete invite
	query = `
    INSERT INTO user_game_membership (user_id, game_id)
    VALUES (:new_player_id, :game_id);
  `
	_, err = db.Sqlite.ExecContext(ctx, query, newPlayerIdArg, gameIdArg)
	if err != nil {
		fmt.Println(" %%% Error while inserting into user_game_membership: ", err)
		return err
	}

	return c.HTML(http.StatusOK, "Joined!")
}

func createQuestion(c echo.Context) error {
	question := c.FormValue("question")
	optionOne := c.FormValue("option-1")
	optionTwo := c.FormValue("option-2")
	optionThree := c.FormValue("option-3")
	optionFour := c.FormValue("option-4")
	gameId := c.QueryParam("game-id")
	askerId := auth.GetFromClaims(auth.UserId, c)

	query := `
  INSERT INTO questions (game_id,asker_id, question_text, option_1, option_2, option_3, option_4)
  VALUES (?, ?, ?, ?, ?, ?, ?);
  `

	_, err := db.Sqlite.Exec(query, gameId, askerId, question, optionOne, optionTwo, optionThree, optionFour)
	if err != nil {
		return c.HTML(http.StatusBadRequest, err.Error())
	}

	return c.HTML(http.StatusOK, `<div id="results"> Hooray</div> <style>#results {font-size: 50px;} </style>`)
}

func getCreateQuestionUI(c echo.Context) error {
	gameId := c.QueryParam("game-id")
	gameIdInt, err := strconv.Atoi(gameId)
	if err != nil {
		fmt.Println("error: provided game-id not convertible to int: ", err)
		return err
	}
	data := struct {
		GameId int
	}{
		GameId: gameIdInt,
	}
	return controllers.RenderTemplate(c, "create-question", data)
}

func getPlayPage(c echo.Context) error {
	gameId := c.Param("game-id")
	gameIdInt, err := strconv.Atoi(gameId)
	if err != nil {
		fmt.Println("game-id param is not convertible to int: ", err)
		return err
	}
	gameIdArg := sql.Named("game_id", gameIdInt)
	fmt.Println(gameIdArg)

	query := `
    SELECT games.name, question_text, questions.id AS question_id, option_1, option_2, option_3, option_4
    FROM questions
    INNER JOIN games
      ON questions.game_id = games.id
    WHERE questions.game_id = :game_id AND DATE(questions.date_created) = DATE('now');
  `

	result := db.Sqlite.QueryRow(query, gameIdArg)

	if err != nil {
		fmt.Println("error querying for question: ", err)
		return err
	}

	var (
		questionRaw    sql.NullString
		questionIdRaw  sql.NullInt64
		gameNameRaw    sql.NullString
		optionOneRaw   sql.NullString
		optionTwoRaw   sql.NullString
		optionThreeRaw sql.NullString
		optionFourRaw  sql.NullString
	)

	err = result.Scan(&gameNameRaw, &questionRaw, &questionIdRaw, &optionOneRaw, &optionTwoRaw, &optionThreeRaw, &optionFourRaw)
	if err != nil {
		fmt.Println("error scanning question data into vars: ", err)
		return err
	}

	type GameData struct {
		GameId      int
		GameName    string
		Question    string
		QuestionId  int64
		OptionOne   string
		OptionTwo   string
		OptionThree string
		OptionFour  string
	}

	data := struct {
		GameData GameData
	}{
		GameData: GameData{
			GameId:      gameIdInt,
			GameName:    gameNameRaw.String,
			Question:    questionRaw.String,
			QuestionId:  questionIdRaw.Int64,
			OptionOne:   optionOneRaw.String,
			OptionTwo:   optionTwoRaw.String,
			OptionThree: optionThreeRaw.String,
			OptionFour:  optionFourRaw.String,
		},
	}

	return controllers.RenderTemplate(c, "gameplay", data)
}

func createAnswer(c echo.Context) error {
	// Get necessary data to insert new answer
	userId := auth.GetFromClaims(auth.UserId, c)
	gameId := c.Param("game-id")
	questionId := c.Param("question-id")
	answer := c.FormValue("answer")

	var optionChosen int

	// convert answer queryParam to an integer
	switch answer {
	case "option-1":
		optionChosen = 1
	case "option-2":
		optionChosen = 2
	case "option-3":
		optionChosen = 3
	case "option-4":
		optionChosen = 4
	default:
		fmt.Println("answer from query param is not a valid option (option-1... option-4). answer provided:", answer)
		return errors.New("invalid option chosen")
	}

	optionChosenArg := sql.Named("option_chosen", optionChosen)
	myUserIdArg := sql.Named("my_user_id", userId)
	gameIdArg := sql.Named("game_id", gameId)
	questionIdArg := sql.Named("question_id", questionId)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}

	defer tx.Rollback()

	query := `
  INSERT INTO answers (game_id, question_id, option_chosen, answerer_id)
  VALUES (:game_id, :question_id, :option_chosen, :my_user_id);
  `

	_, err = db.Sqlite.ExecContext(ctx, query, myUserIdArg, optionChosenArg, gameIdArg, questionIdArg)
	if err != nil {
		cancel()
		tx.Rollback()
		fmt.Println("error inserting answer into db: ", err)
		return err
	}

	fmt.Println("^^^^^^^^ gameId:", gameId)
	fmt.Println("^^^^^^^^ answer: ", answer)

	query = `
	   WITH vote_counts AS (
	   SELECT COUNT(*) AS votes, option_chosen, question_id
	   FROM answers
	   JOIN questions
	       ON questions.id = answers.question_id
	       AND
	       DATE(questions.date_created) = DATE('now')
	   WHERE answers.game_id = :game_id AND answers.question_id = :question_id
	   GROUP BY answers.option_chosen
	   )
	   SELECT  votes, vote_counts.option_chosen,
	     CASE 
	         WHEN answers.option_chosen = 1
	         THEN questions.option_1
	         WHEN answers.option_chosen = 2
	         THEN questions.option_2
	         WHEN answers.option_chosen = 3
	         THEN questions.option_3
	         ELSE questions.option_4
	     END AS answer_text
	   FROM answers
	   JOIN vote_counts
	     ON vote_counts.option_chosen = answers.option_chosen
	        AND
	        answers.question_id = vote_counts.question_id
	   JOIN questions ON questions.id = answers.question_id
      GROUP BY vote_counts.option_chosen
      ORDER BY votes;
  `

	// Determine the number of votes for each option
	// query = `
	//    WITH vote_counts AS (
	//    SELECT COUNT(*) AS votes, option_chosen, question_id
	//    FROM answers
	//    JOIN questions
	//        ON questions.id = answers.question_id
	//        AND
	//        DATE(questions.date_created) = DATE('now')
	//    WHERE answers.game_id = :game_id AND answers.question_id = :question_id
	//    GROUP BY answers.option_chosen
	//    )
	//    SELECT answerer_id, votes, answers.option_chosen,
	//      CASE
	//          WHEN answers.option_chosen = 1
	//          THEN questions.option_1
	//          WHEN answers.option_chosen = 2
	//          THEN questions.option_2
	//          WHEN answers.option_chosen = 3
	//          THEN questions.option_3
	//          ELSE questions.option_4
	//      END AS answer_text
	//    FROM answers
	//    JOIN vote_counts
	//      ON vote_counts.option_chosen = answers.option_chosen
	//         AND
	//         answers.question_id = vote_counts.question_id
	//    JOIN questions ON questions.id = answers.question_id;
	//  `

	rows, err := db.Sqlite.QueryContext(ctx, query, gameIdArg, questionIdArg)
	if err != nil {
		fmt.Println("*** error querying db for question results: ", err)
		tx.Rollback()
		cancel()
		return err
	}

	type AnswerStats struct {
		Option     int64
		Votes      int64
		AnswerText string
	}

	var answerStats []AnswerStats

	for rows.Next() {
		var (
			optionRaw     sql.NullInt64
			votesRaw      sql.NullInt64
			answerTextRaw sql.NullString
		)

		err := rows.Scan(&votesRaw, &optionRaw, &answerTextRaw)
		if err != nil {
			fmt.Println("error scanning answer votes into vars: ", err)
			cancel()
			tx.Rollback()
			return err
		}

		answerStats = append(answerStats, struct {
			Option     int64
			Votes      int64
			AnswerText string
		}{
			Option:     optionRaw.Int64,
			Votes:      votesRaw.Int64,
			AnswerText: answerTextRaw.String,
		})
	}

	data := struct {
		Data []AnswerStats
	}{
		Data: answerStats,
	}

	fmt.Println("%%%% answerStats: ", answerStats, " %%%%%%")

	return controllers.RenderTemplate(c, "results", data)
}
