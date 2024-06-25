WITH todays_answers AS (
  SELECT answer_text, users.username, users.id,
    (
      SELECT COUNT(*)
      FROM votes v
      WHERE v.answer_id IN (
          SELECT questions.id
          FROM questions
          WHERE DATE(questions.date_created) = DATE('now')
        )
    ) AS votes
  FROM answers
  JOIN users
    ON answers.answerer_id = users.id
  WHERE (
      SELECT questions.id
      FROM questions
      WHERE DATE(questions.date_created) = DATE('now')
      ) = answers.question_id
  JOIN votes
    ON votes.answer_id = answers.id
)
