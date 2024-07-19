package games

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func CreateQuestion(c echo.Context, gameId int, askerId int, questionText string, optionOne int, optionTwo int, optionThree int, optionFour int) error {
	query := `
  INSERT INTO questions (game_id,asker_id, question_text, option_1, option_2, option_3, option_4)
  VALUES (?, ?, ?, ?, ?, ?, ?);
  `

	_, err := db.Sqlite.Exec(query, gameId, askerId, questionText, optionOne, optionTwo, optionThree, optionFour)
	if err != nil {
		return c.HTML(http.StatusBadRequest, err.Error())
	}
	return nil
}

func GetCurrentQuestionId(tx *sql.Tx, gameId int) (int64, error) {
	fail := func(err error) (int64, error) {
		return 0, fmt.Errorf("getCurrentQuestionId: %v", err)
	}

	gameIdArg := sql.Named("game_id", gameId)

	query := `
      SELECT id
      FROM questions
      WHERE game_id = :game_id
        AND DATE(date_created) = DATE('now');
  `

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	row := tx.QueryRowContext(ctx, query, gameIdArg)

	var questionIdRaw sql.NullInt64

	err := row.Scan(&questionIdRaw)
	if err != nil {
		tx.Rollback()
		return fail(err)
	}

	return questionIdRaw.Int64, nil
}
