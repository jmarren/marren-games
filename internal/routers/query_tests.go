package routers

import "reflect"

type routeConfig struct {
	path        string
	method      string
	query       string
	queryParams []ParamConfig
}

type ParamConfig struct {
	Name string
	Type reflect.Kind
}

type routeConfigs []*routeConfig

func CreateNewRouteConfigs(r []routeConfig) routeConfigs {
	var routeConfigs routeConfigs
	for _, routeConfig := range r {
		routeConfigs = append(routeConfigs, &routeConfig)
	}
	return routeConfigs
}

func CreateNewRouteConfig() *routeConfig {
	return &routeConfig{}
}

func GetRouteConfigs() routeConfigs {
	routeConfigs := CreateNewRouteConfigs(
		[]routeConfig{
			{
				path:   "/get-username-with-id",
				method: "GET",
				query:  `SELECT username FROM users WHERE id = ?`,
				queryParams: []ParamConfig{
					{Name: "id", Type: reflect.Int},
				},
			},
			{
				path:   "/questions-by-user-id",
				method: "GET",
				query: `SELECT question_text, users.username, date_created
                FROM questions
                INNER JOIN users
                  ON users.id = questions.asker_id
                WHERE questions.asker_id = ?;`,
				queryParams: []ParamConfig{
					{Name: "id", Type: reflect.Int},
				},
			},
			{
				path:   "/create-question",
				method: "POST",
				query: `INSERT INTO questions (asker_id, question_text)
                VALUES (?,?);`,
				queryParams: []ParamConfig{
					{Name: "id", Type: reflect.Int},
					{Name: "text", Type: reflect.String},
				},
			},
			{
				path:        "/todays-question",
				method:      "GET",
				query:       `SELECT question_text, questions.id FROM questions WHERE DATE(CURRENT_TIMESTAMP) = DATE(questions.date_created);`,
				queryParams: []ParamConfig{},
			},
			{
				path:   "/answer-to-todays-question",
				method: "GET",
				query: `SELECT answers.answer_text
                FROM answers
                WHERE answers.answerer_id = ?
                AND answers.question_id = (SELECT * FROM todays_question_id)`,
				queryParams: []ParamConfig{
					{Name: "user_id", Type: reflect.Int},
				},
			},
			{
				path:   "/answer-to-todays-question",
				method: "POST",
				query: `INSERT INTO answers (answerer_id, question_id, answer_text)
                VALUES (?, (SELECT questions.id FROM questions WHERE DATE(CURRENT_TIMESTAMP) = DATE(questions.date_created)), ?);`,
				queryParams: []ParamConfig{
					{Name: "user_id", Type: reflect.Int},
					{Name: "answer_text", Type: reflect.String},
				},
			},
			{
				path:   "/vote-for-answer",
				method: "POST",
				query: `INSERT INTO votes (voter_id, question_id)
                VALUES (?, (SELECT * FROM todays_question_id));`,
				queryParams: []ParamConfig{
					{Name: "user_id", Type: reflect.Int},
				},
			},
			{
				path:   "/all-answers-to-todays-question",
				method: "GET",
				query: `SELECT answer_text
                FROM answers
                WHERE question_id = (SELECT * FROM todays_question_id);`,
				queryParams: []ParamConfig{},
			},
			{
				path:   "/check-if-todays-question-answered",
				method: "GET",
				query: `WITH answer_exists AS (
                  SELECT answerer_id
                  FROM answers
                  WHERE answerer_id = ?
                  AND question_id = (SELECT * FROM todays_question_id)
                )
                SELECT
                  CASE
                    WHEN EXISTS (SELECT ? FROM answer_exists)
                      THEN 1
                    ELSE 0
                  END AS todays_question_answered;`,
				queryParams: []ParamConfig{
					{Name: "user_id", Type: reflect.Int},
					{Name: "user_id", Type: reflect.Int},
					{Name: "user_id", Type: reflect.Int},
				},
			},
      {
        path: 
      },
			{
				path:   "/profile",
				method: "GET",
				query: `SELECT
                  users.username, questions.asker_id, questions.question_text, answers.answer_text, votes.question_id
                FROM
                  users
                WHERE
                  username = ?
                JOIN
                  questions
                    ON
                       DATE(questions.date_created) = DATE(CURRENT_TIMESTAMP)
                JOIN
                  answers
                    ON
                      questions.id = answers.question_id
                JOIN
                  votes
                    ON
                      votes.voter_id = users.id
                VALUES (?)`,
				queryParams: []ParamConfig{
					{Name: "username", Type: reflect.String},
				},
			},
		})
	return routeConfigs
}
