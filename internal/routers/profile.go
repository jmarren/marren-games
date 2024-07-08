package routers

import (
	"database/sql"
	"reflect"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/labstack/echo/v4"
)

func GetProfileRoutes() []*RouteConfig {
	return CreateNewRouteConfigs(
		[]RouteConfig{
			{
				path:   "/",
				method: GET,
				claimArgConfigs: []ClaimArgConfig{
					{claim: auth.Username, Type: reflect.String},
				},
				urlQueryParamArgConfigs: []UrlQueryParamArgConfig{},
				urlPathParamArgConfigs:  []UrlPathParamArgConfig{},
				withQueries:             []string{},
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
						Username   string `json:"answerer_username"`
						AnswerText string `json:"answer_text" `
						Votes      int    `json:"votes"`
					}
				}{},
			},
		})
}

func ProfileRouter(r *echo.Group) {
	r.GET("/transition/:target-page", func(c echo.Context) error {
		targetPage := c.Param("target-page")
		dataStruct := struct {
			TargetPage string
		}{
			TargetPage: targetPage,
		}

		data := TemplateData{
			Data: dataStruct,
		}
		return controllers.RenderTemplate(c, "slide-out-to-right", data)
	})
}
