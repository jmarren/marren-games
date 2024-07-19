package games

import (
	"context"
	"database/sql"
	"fmt"
)

func getAnswerStats(ctx *context.Context, tx *sql.Tx, questionIdArg sql.NamedArg, gameIdArg sql.NamedArg) ([]AnswerStats, error) {
	query := `
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

	rows, err := tx.QueryContext(*ctx, query, gameIdArg, questionIdArg)
	if err != nil {
		fmt.Println("*** error querying db for question results: ", err)
		tx.Rollback()
		return nil, err
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
			tx.Rollback()
			return nil, err
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

	return answerStats, nil
}
