package games

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

func ResultsRouter(r *echo.Group) {
	// TODO
	r.GET("", GetGameResults)
	r.POST("", createAnswer)
	// r.PUT("", updateGameResults)
}

func createAnswer(c echo.Context) error {
	// Get necessary data to insert new answer
	userId := auth.GetFromClaims(auth.UserId, c)
	gameId := c.Param("game-id")
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

	// convert data to sql.NamedArg type
	optionChosenArg := sql.Named("option_chosen", optionChosen)
	myUserIdArg := sql.Named("my_user_id", userId)
	gameIdArg := sql.Named("game_id", gameId)

	// create context for query
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	// begin Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}

	defer tx.Rollback()

	// Get todays question id
	query := `
    SELECT id
    FROM questions
    WHERE game_id = :game_id
    AND DATE(date_created) = DATE('now');
  `

	var questionIdRaw sql.NullInt64

	row := tx.QueryRowContext(ctx, query, gameIdArg)

	err = row.Scan(&questionIdRaw)
	if err != nil {
		return fmt.Errorf("error scanning questionId into var")
	}

	// convert todays question id to sql.NamedArg
	questionId := questionIdRaw.Int64

	questionIdArg := sql.Named("question_id", questionId)

	// query to insert answer
	query = `
  INSERT INTO answers (game_id, question_id, option_chosen, answerer_id)
  SELECT :game_id, :question_id, :option_chosen, :my_user_id
  `

	// perform query to insert answer
	_, err = tx.ExecContext(ctx, query, myUserIdArg, optionChosenArg, gameIdArg, questionIdArg)
	if err != nil {
		tx.Rollback()
		fmt.Println("error: resultsRouter, createAnswer(), error inserting answer into db: ", err)
		return err
	}

	// query to update scores for other players
	query = `
    WITH users_to_increment AS (
      SELECT answerer_id 
        FROM answers
        WHERE answers.option_chosen = :option_chosen
              AND answers.game_id = :game_id
              AND answers.question_id = :question_id
              AND answers.answerer_id != :my_user_id
    )
    UPDATE scores
    SET score = (score + 1)
    WHERE scores.user_id IN users_to_increment;
  `
	_, err = tx.ExecContext(ctx, query, myUserIdArg, optionChosenArg, gameIdArg, questionIdArg)
	if err != nil {
		cancel()
		tx.Rollback()
		fmt.Println("error inserting answer into db: ", err)
		return err
	}

	// update score for answerer
	query = `
    UPDATE scores
    SET score = score + (
          SELECT COUNT(*)
          FROM answers
          WHERE answers.option_chosen
            AND answers.game_id = :game_id
            AND answers.question_id = :question_id
    )
    WHERE scores.user_id = :my_user_id;
  `
	_, err = tx.ExecContext(ctx, query, myUserIdArg, optionChosenArg, gameIdArg, questionIdArg)
	if err != nil {
		cancel()
		tx.Rollback()
		fmt.Println("error: resultsRouter, createAnswer: error inserting answer into db: ", err)
		return err
	}

	tx.Commit()

	return GetGameResults(c)
}

// TODO
func GetGameResults(c echo.Context) error {
	// get game id from query params
	gameId, err := strconv.Atoi(c.Param("game-id"))
	if err != nil {
		return fmt.Errorf("game-id not convertible to int: %v", err)
	}

	// convert game id to sql.NamedArg
	gameIdArg := sql.Named("game_id", gameId)

	// create context for query
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	// begin Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}

	fmt.Printf("\n \033[31m gameId: %v  \033[0m \n", gameId)

	defer tx.Rollback()

	query := `
	   WITH vote_counts AS (
	   SELECT COUNT(*) AS votes, option_chosen, question_id
	   FROM answers
	   JOIN questions
	       ON questions.id = answers.question_id
	       AND
	       DATE(questions.date_created) = DATE('now')
	   WHERE answers.game_id = :game_id AND answers.question_id = (
          SELECT id
          FROM questions
          WHERE game_id = :game_id
            AND DATE(date_created) = DATE('now')
        )
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

	rows, err := tx.QueryContext(ctx, query, gameIdArg)
	if err != nil {
		fmt.Println("*** error querying db for question results: ", err)
		tx.Rollback()
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

	query = `
    SELECT score, users.username
    FROM scores
    JOIN users ON users.id = scores.user_id
    WHERE game_id = :game_id
    ORDER BY score;
    `

	rows, err = db.Sqlite.QueryContext(ctx, query, gameIdArg)
	if err != nil {
		fmt.Println("*** error querying db for question results: ", err)
		tx.Rollback()
		return err
	}

	type UserScore struct {
		Username string
		Score    int64
	}

	var scoreboardData []UserScore

	for rows.Next() {
		var (
			scoreRaw    sql.NullInt64
			usernameRaw sql.NullString
		)
		err := rows.Scan(&scoreRaw, &usernameRaw)
		if err != nil {
			fmt.Println("error scanning answer scoreboard data into vars: ", err)
			tx.Rollback()
			return err
		}
		scoreboardData = append(scoreboardData, UserScore{
			Username: usernameRaw.String,
			Score:    scoreRaw.Int64,
		})

	}

	data := struct {
		AnswersData    []AnswerStats
		ScoreboardData []UserScore
	}{
		AnswersData:    answerStats,
		ScoreboardData: scoreboardData,
	}

	fmt.Println("%%%% answerStats: ", answerStats, " %%%%%%")

	return controllers.RenderTemplate(c, "results", data)
}

//
// func updateGameResults(c echo.Context) error {
// }
