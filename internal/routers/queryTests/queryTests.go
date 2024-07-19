package queryTests

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"

	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

type QueryConfig struct {
	path        string
	method      RouteMethod
	withQueries []string
	query       string
	params      []ParamConfig
}

type ParamConfig struct {
	Name string
	Type reflect.Kind
}

var QueryConfigs = []QueryConfig{
	{
		path:   "/users",
		method: GET,
		query: `SELECT users.username, users.email,
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
            SELECT json_group_array(
              json_object(
                  'answerer_username', abv.answerer_username,
                  'answer_text', abv.answer_text,
                  'votes', abv.total_votes
                  )
                )
                FROM answers_by_votes abv
            ) AS answers
            FROM users
            WHERE users.username = :Username;`,
		params: []ParamConfig{
			{Name: "Username", Type: reflect.String},
		},
	},
	{
		path:   "/all-answers",
		method: GET,
		query: `SELECT answer_text, users.username, users.id
            FROM answers
            JOIN users
              ON answers.answerer_id = users.id
            WHERE (
                SELECT questions.id
                FROM questions
                WHERE DATE(questions.date_created) = DATE('now')
                ) = answers.question_id
`,
		params: []ParamConfig{},
	},
	{
		path:   "/view/todays_answers",
		method: GET,
		query: `
            SELECT
                answers.answer_text,
                answers.id AS answer_id,
                users.id AS answerer_id,
                users.username AS answerer_username,
                questions.id AS question_id
            FROM
                answers
            JOIN
                users
              ON
                users.id = answer_id
            JOIN
                questions
                  ON question_id = answers.question_id
            WHERE
                DATE(answers.date_created) = DATE('now')
                AND DATE(questions.date_created) = DATE('now');
            `,
		withQueries: []string{},
		params:      []ParamConfig{},
	},
	{
		path:        "/wq/all-answers",
		method:      GET,
		withQueries: []string{"todays_answers"},
		query:       `SELECT * FROM todays_answers;`,
		params:      []ParamConfig{},
	},
	{
		path:        "/view/answers_by_votes",
		method:      GET,
		withQueries: []string{},
		query:       `SELECT * FROM answers_by_votes;`,
		params:      []ParamConfig{},
	},
}

func QueryTests(group *echo.Group) {
	for _, queryConfig := range QueryConfigs {
		switch queryConfig.method {
		case GET:
			group.GET(queryConfig.path,

				func(c echo.Context) error {
					// convert params to the type specified in config
					var params []interface{}

					for _, paramConfig := range queryConfig.params {
						paramValue := c.QueryParam(paramConfig.Name)
						convertedValue, err := ConvertType(paramValue, paramConfig.Type)
						if err != nil {
							errorMessage := fmt.Errorf("**** error converting type: %s\n| parameter name: %s\n| parameter value: %s\n| paramConfig.Type: %s ", err, paramConfig.Name, paramValue, paramConfig.Type)
							fmt.Println(errorMessage)
							return c.String(http.StatusBadRequest, error.Error(errorMessage))
						}
						namedParam := sql.Named(paramConfig.Name, convertedValue)
						params = append(params, namedParam)
					}
					fmt.Println(params)

					query := GetFullQuery(queryConfig.query, queryConfig.withQueries)

					// Perform Query
					outputRows, err := db.Sqlite.Query(query, params...)
					if err != nil {
						fmt.Println(err)
						return err
					}

					cols, err := outputRows.Columns()
					if err != nil {
						fmt.Println(err)
						return err
					}
					colLen := len(cols)
					vals := make([]interface{}, colLen)
					valPtrs := make([]interface{}, colLen)

					response := ""

					for outputRows.Next() {
						for i := range cols {
							valPtrs[i] = &vals[i]
						}
						err := outputRows.Scan(valPtrs...)
						if err != nil {
							fmt.Println(err)
							return err
						}
						for i, col := range cols {
							val := vals[i]

							b, ok := val.([]byte)
							var v interface{}
							if ok {
								v = string(b)
							} else {
								v = val
							}
							fmt.Println(col, v)
							response = fmt.Sprintf("%v\n %v: %v", response, col, v)
						}
						response = fmt.Sprintf("%v\n-----------", response)
					}

					return c.String(http.StatusOK, response)
				})
		}
	}
}
