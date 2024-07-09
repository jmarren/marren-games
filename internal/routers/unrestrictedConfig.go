////////////
////////////  This file is used for testing various queries on the sqlite3/libsql database.
////////////  The RouteConfig slice is looped over in ./unrestricted.go in order to
////////////  declare a route for each item in the slice.
////////////
////////////  All routes are served at /query{RouteConfig.path}
////////////
////////////  The query for each RouteConfig is executed and the data is returned.
////////////  The withQueries contains the name of a file in ../sql (or /internal/sql) directory
////////////  contains a 'WITH' clause which will be appended to the query in order to provide
////////////  an abstraction for some commonly accessed piece of data.
////////////
////////////  NOTE: Any named parameters prepended with a ':' (ie :user_id) within a query should
////////////  be included in urlQueryParamArgConfigs. queryParams should also include any named parameters
////////////  used by any corresponding withQueries provided.

package routers

import (
	"database/sql"
	"reflect"
)

type RouteConfig struct {
	path                    string
	method                  RouteMethod
	withQueries             []string
	query                   string
	claimArgConfigs         []ClaimArgConfig
	urlPathParamArgConfigs  []UrlPathParamArgConfig
	urlQueryParamArgConfigs []UrlQueryParamArgConfig
	partialTemplate         string
	typ                     interface{}
}

type RouteConfigs []*RouteConfig

func CreateNewRouteConfigs(r []RouteConfig) RouteConfigs {
	var RouteConfigs RouteConfigs
	for _, RouteConfig := range r {
		RouteConfigs = append(RouteConfigs, &RouteConfig)
	}
	return RouteConfigs
}

func CreateNewRouteConfig() *RouteConfig {
	return &RouteConfig{}
}

func GetRouteConfigs() RouteConfigs {
	RouteConfigs := CreateNewRouteConfigs(
		[]RouteConfig{
			{
				path:   "/get-username-with-id",
				method: GET,
				query:  `SELECT username FROM users WHERE id = :user_id`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "user_id", Type: reflect.Int},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
			},
			{
				path:   "/questions-by-user-id",
				method: GET,
				query: `SELECT question_text, users.username, date_created
                      FROM questions
                      INNER JOIN users
                      ON users.id = questions.asker_id
                      WHERE questions.asker_id = :user_id;`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "user_id", Type: reflect.Int},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
			},
			{
				path:   "/create-question",
				method: POST,
				query: `INSERT INTO questions (asker_id, question_text)
                VALUES ($user_id,$question_text);`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "user_id", Type: reflect.Int},
					{Name: "question_text", Type: reflect.String},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
			},
			{
				path:                    "/todays-question",
				method:                  GET,
				query:                   `SELECT * FROM todays_question;`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
			},
			{
				path:        "/answer-to-todays-question",
				method:      GET,
				withQueries: []string{"todays_answer_by_user_id"},
				query:       `SELECT answer_text FROM todays_answer;`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "user_id", Type: reflect.Int},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
			},
			{
				path:   "/answer-to-todays-question",
				method: POST,
				query: `INSERT INTO answers (answerer_id, question_id, answer_text)
        VALUES (:user_id, (SELECT questions.id FROM questions WHERE DATE(CURRENT_TIMESTAMP) = DATE(questions.date_created)), :answer_text);`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "user_id", Type: reflect.Int},
					{Name: "answer_text", Type: reflect.String},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
			},
			{
				path:   "/vote-for-answer",
				method: POST,
				query: `INSERT INTO votes (voter_id, question_id, answer_id)
        VALUES (:user_id, (SELECT * FROM todays_question_id), :answer_id);`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "user_id", Type: reflect.Int},
					{Name: "answer_id", Type: reflect.Int},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
			},
			{
				path:   "/all-answers-to-todays-question",
				method: GET,
				query: `SELECT answer_text
                FROM answers
                WHERE question_id = (SELECT * FROM todays_question_id);`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
			},
			{
				path:   "/check-if-todays-question-answered",
				method: GET,
				query: `WITH answer_exists AS (
                  SELECT answerer_id
                  FROM answers
                    WHERE answerer_id = :user_id
                    AND question_id = (SELECT * FROM todays_question_id)
                )
                SELECT
                  CASE
                    WHEN EXISTS (SELECT :user_id FROM answer_exists)
                      THEN 1
                    ELSE 0
                  END AS todays_question_answered;`,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "user_id", Type: reflect.Int},
					{Name: "user_id", Type: reflect.Int},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
			},
			{
				path:   "/todays-answers-2",
				method: GET,
				query: `SELECT *
                FROM todays_answers;`,
				withQueries:             []string{},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				claimArgConfigs:         []ClaimArgConfig{},
				partialTemplate:         "profile",
				typ: struct {
					AnswerText       sql.NullString
					AnswererID       sql.NullInt64
					AnswererUsername sql.NullString
					QuestionID       sql.NullInt64
				}{},
			},
			{
				path:   "/all-answers",
				method: GET,
				query: `SELECT *
                FROM answers;`,
				withQueries:             []string{},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				claimArgConfigs:         []ClaimArgConfig{},
				partialTemplate:         "profile",
				typ: struct {
					AnswerId    sql.NullInt64
					AnswerText  sql.NullString
					DateCreated sql.NullString
					QuestionId  sql.NullInt64
					AnswererId  sql.NullInt64
				}{},
			},
			{
				path:   "/profile",
				method: GET,
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{
					{Name: "Username", Type: reflect.String},
				},
				urlPathParamArgConfigs: []UrlPathParamArgConfig{},
				withQueries:            []string{},
				query: `
          SELECT users.username, users.email,
            CASE
                WHEN users.id = (
                    SELECT questions.asker_id
                    FROM questions
                    WHERE DATE(questions.date_created) = DATE('now')
                  ) THEN 1
                ELSE 0
            END AS is_asker,
            CASE
                WHEN (SELECT answers.answer_text
                  FROM answers
                  WHERE answers.answerer_id = users.id
                  AND answers.question_id = (
                    SELECT questions.id
                    FROM questions
                    WHERE DATE(questions.date_created) = DATE('now')
                  )
                ) IS NOT NULL THEN 1
            ELSE 0
            END AS answered_today,
            (
              SELECT q.question_text
              FROM questions q
              WHERE DATE(q.date_created) = DATE('now')
              LIMIT 1
            ) AS todays_question_text,
          (
            SELECT
                json_group_array (
                  json_object(
                      'answerer_username', abv.answerer_username,
                      'answer_text', abv.answer_text,
                      'votes', abv.total_votes
                      )
                  ) FROM answers_by_votes abv
                ) AS answers
            FROM users
            WHERE users.username = :Username;
        `,
				partialTemplate: "profile",
				typ: struct {
					Username      sql.NullString
					Email         sql.NullString
					IsAsker       sql.NullInt64
					AnsweredToday sql.NullInt64
					QuestionText  sql.NullString
					Answers       []struct {
						Username   string `json:"answerer_username"`
						AnswerText string `json:"answer_text" `
						Votes      int    `json:"votes"`
					}
				}{},
			},
			{
				path:   "/create-account",
				method: "POST",
			},
		})
	return RouteConfigs
}
