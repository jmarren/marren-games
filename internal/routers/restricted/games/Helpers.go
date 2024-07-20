package games

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmarren/marren-games/internal/db"
)

func GetTodaysQuestionId(gameId int) (int64, error) {
	gameIdArg := sql.Named("game_id", gameId)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	// begin Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return 0, err
	}

	defer tx.Rollback()

	// Get todays question id
	query := `
    SELECT id
    FROM questions
    WHERE game_id = :game_id
    AND DATE(date_created) = DATE('now')
  `

	var questionIdRaw sql.NullInt64

	row := tx.QueryRowContext(ctx, query, gameIdArg)

	err = row.Scan(&questionIdRaw)
	if err != nil {
		return 0, fmt.Errorf("error scanning questionId into var")
	}

	// convert todays question id to sql.NamedArg
	questionId := questionIdRaw.Int64
	return questionId, nil
}
