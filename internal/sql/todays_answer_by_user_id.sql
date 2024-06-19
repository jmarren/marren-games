WITH todays_answer AS (
    SELECT answers.answer_text, users.id
    FROM answers
    WHERE answers.answerer_id = :user_id
    AND answers.question_id = (SELECT * FROM todays_question_id)
    LIMIT 1
    )
