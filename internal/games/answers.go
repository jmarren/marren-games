package games

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func CreateAnswer(c echo.Context, optionChosen int, userId int, gameId int, questionId int) error {
	fail := func(err error) error {
		return fmt.Errorf("createAnswer: %v ", err)
	}
	// Get necessary data to insert new answer
	/*
	     userId := auth.GetFromClaims(auth.UserId, c)
	   	gameId := c.Param("game-id")
	   	questionId := c.Param("question-id")
	   	gameIdInt, err := strconv.Atoi(gameId)
	   	answer := c.FormValue("answer")
	   	if err != nil {
	   		// fmt.Println("game-id param not convertible to int")
	   		// return err
	   		return fail(err)
	   	}
	   	questionIdInt, err := strconv.Atoi(questionId)
	   	if err != nil {
	   		fmt.Println("game-id param not convertible to int")
	   		return err
	   	}

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
	*/

	optionChosenArg := sql.Named("option_chosen", optionChosen)
	myUserIdArg := sql.Named("my_user_id", userId)
	gameIdArg := sql.Named("game_id", gameId)
	questionIdArg := sql.Named("question_id", questionId)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}

	defer tx.Rollback()

	query := `
  INSERT INTO answers (game_id, question_id, option_chosen, answerer_id)
  VALUES (:game_id, :question_id, :option_chosen, :my_user_id);
  `

	_, err = tx.ExecContext(ctx, query, myUserIdArg, optionChosenArg, gameIdArg, questionIdArg)
	if err != nil {
		tx.Rollback()
		fmt.Println("error inserting answer into db: ", err)
		return err
	}

	fmt.Println("^^^^^^^^ gameId:", gameId)
	fmt.Println("^^^^^^^^ answer: ", answer)

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
		tx.Rollback()
		fmt.Println("error inserting answer into db: ", err)
		return err
	}

	scoreboardData, err := getUserScores(&ctx, tx, gameIdInt)
	if err != nil {
		fmt.Println("error getting user scores")
		return err
	}

	answerStats, err := getAnswerStats(&ctx, tx, questionIdArg, gameIdArg)
	if err != nil {
		return fail(err)
	}

	data := struct {
		AnswersData    []AnswerStats
		ScoreboardData []UserScore
	}{
		AnswersData:    answerStats,
		ScoreboardData: scoreboardData,
	}

	tx.Commit()
	fmt.Println("%%%% answerStats: ", answerStats, " %%%%%%")

	return controllers.RenderTemplate(c, "results", data)
}
