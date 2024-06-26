package routers

import (
	"database/sql"
	"reflect"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
)

type PageData struct {
	title    string
	template controllers.TemplateName
	data     DataInterface
}

type DataInterface interface {
	executeQuery(string)
}

type RouteMethod string

const (
	GET    RouteMethod = "GET"
	POST   RouteMethod = "POST"
	PUT    RouteMethod = "PUT"
	DELETE RouteMethod = "DELETE"
)

type UrlParam string

const (
	user_id  UrlParam = "user_id"
	username UrlParam = "username"
)

type QueryParam interface {
	getValue()
}

type ClaimArgConfig struct {
	claim auth.ClaimsType
	Type  reflect.Kind
}

type UrlParamArgConfig struct {
	urlParam UrlParam
	Type     reflect.Kind
}

type Vote struct {
	voterId       int
	voterUsername string
}

func GetRestrictedRouteConfigs() []*RouteConfig {
	return CreateNewRouteConfigs(
		[]RouteConfig{
			{
				path:   "/profile",
				method: GET,
				claimArgConfigs: []ClaimArgConfig{
					{claim: auth.Username, Type: reflect.String},
				},
				urlParamArgConfigs: []UrlParamArgConfig{},
				withQueries:        []string{},
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
            SELECT json_array(
            json_object(
                'answerer_username', abv.answerer_username,
                'answer_text', abv.answer_text,
                'votes', abv.total_votes
                )
              )
                FROM answers_by_votes abv
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
					Answers       struct {
						Username   sql.NullString
						AnswerText sql.NullString
						votes      sql.NullInt64
					}
				}{},
			},
		},
	)
}

// (
//   SELECT GROUP_CONCAT(
//   '{' ||
//         'username: ' || abv.answerer_username ||
//         ', answer: ' || abv.answer_text ||
//         ', votes: ' || abv.total_votes
//   || '}', ',')
//     FROM answers_by_votes abv
// ) AS answers
