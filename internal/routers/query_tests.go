////////////
////////////  This file is used for testing various queries on the sqlite3/libsql database.
////////////  The routeConfig slice is looped over in ./unrestricted.go in order to
////////////  declare a route for each item in the slice.
////////////
////////////  All routes are served at /query{routeConfig.path}
////////////
////////////  The query for each routeConfig is executed and the data is returned.
////////////  The withQuery contains the name of a file in ../sql (or /internal/sql) directory
////////////  contains a 'WITH' clause which will be appended to the query in order to provide
////////////  an abstraction for some commonly accessed piece of data.
////////////
////////////  NOTE: Any named parameters prepended with a ':' (ie :user_id) within a query should
////////////  be included in urlParamArgConfigs. queryParams should also include any named parameters
////////////  used by any corresponding withQuery provided.

package routers

import (
	"database/sql"
	"reflect"

	"github.com/jmarren/marren-games/internal/db"
)

type routeConfig struct {
	path               string
	method             string
	withQuery          string
	query              string
	claimArgConfigs    []ClaimArgConfig
	urlParamArgConfigs []UrlParamArgConfig
	createNewSlice     func() db.RowContainer
	typ                reflect.Type
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

type AnswersStruct struct {
	Answers []Answer
}

func CreateAnswer() db.RowContainer {
	return &Answer{
		AnswerText:       sql.NullString{},
		AnswererID:       sql.NullInt64{},
		AnswererUsername: sql.NullString{},
		QuestionID:       sql.NullInt64{},
	}
}

type Answer struct {
	AnswerText       sql.NullString `xml:"Answer_text"`
	AnswererID       sql.NullInt64  `xml:"Answerer_id"`
	AnswererUsername sql.NullString `xml:"Answerer_username"`
	QuestionID       sql.NullInt64  `xml:"Question_id"`
}

func (a *Answer) GetPtrs() []interface{} {
	return []interface{}{&a.AnswerText, &a.AnswererID, &a.AnswererUsername, &a.QuestionID}
}

// var AnswerSlice []Answer

func GetRouteConfigs() routeConfigs {
	routeConfigs := CreateNewRouteConfigs(
		[]routeConfig{
			{
				path:   "/get-username-with-id",
				method: "GET",
				query:  `SELECT username FROM users WHERE id = :user_id`,
				urlParamArgConfigs: []UrlParamArgConfig{
					{urlParam: "user_id", Type: reflect.Int},
				},
			},
			{
				path:   "/questions-by-user-id",
				method: "GET",
				query: `SELECT question_text, users.username, date_created
                      FROM questions
                      INNER JOIN users
                      ON users.id = questions.asker_id
                      WHERE questions.asker_id = :user_id;`,
				urlParamArgConfigs: []UrlParamArgConfig{
					{urlParam: "user_id", Type: reflect.Int},
				},
			},
			{
				path:   "/create-question",
				method: "POST",
				query: `INSERT INTO questions (asker_id, question_text)
                VALUES ($user_id,$question_text);`,
				urlParamArgConfigs: []UrlParamArgConfig{
					{urlParam: "user_id", Type: reflect.Int},
					{urlParam: "question_text", Type: reflect.String},
				},
			},
			{
				path:               "/todays-question",
				method:             "GET",
				query:              `SELECT * FROM todays_question;`,
				urlParamArgConfigs: []UrlParamArgConfig{},
			},
			{
				path:      "/answer-to-todays-question",
				method:    "GET",
				withQuery: "todays_answer_by_user_id",
				query:     `SELECT answer_text FROM todays_answer;`,
				urlParamArgConfigs: []UrlParamArgConfig{
					{urlParam: "user_id", Type: reflect.Int},
				},
			},
			{
				path:   "/answer-to-todays-question",
				method: "POST",
				query: `INSERT INTO answers (answerer_id, question_id, answer_text)
        VALUES (:user_id, (SELECT questions.id FROM questions WHERE DATE(CURRENT_TIMESTAMP) = DATE(questions.date_created)), :answer_text);`,
				urlParamArgConfigs: []UrlParamArgConfig{
					{urlParam: "user_id", Type: reflect.Int},
					{urlParam: "answer_text", Type: reflect.String},
				},
			},
			{
				path:   "/vote-for-answer",
				method: "POST",
				query: `INSERT INTO votes (voter_id, question_id, answer_id)
        VALUES (:user_id, (SELECT * FROM todays_question_id), :answer_id);`,
				urlParamArgConfigs: []UrlParamArgConfig{
					{urlParam: "user_id", Type: reflect.Int},
					{urlParam: "answer_id", Type: reflect.Int},
				},
			},
			{
				path:   "/all-answers-to-todays-question",
				method: "GET",
				query: `SELECT answer_text
                FROM answers
                WHERE question_id = (SELECT * FROM todays_question_id);`,
				urlParamArgConfigs: []UrlParamArgConfig{},
				// dataContainer:      &AnswersStruct{},
			},
			{
				path:   "/check-if-todays-question-answered",
				method: "GET",
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
				urlParamArgConfigs: []UrlParamArgConfig{
					{urlParam: "user_id", Type: reflect.Int},
					{urlParam: "user_id", Type: reflect.Int},
				},
			},
			{
				path:   "/todays-answers",
				method: "GET",
				query: `SELECT *
                FROM todays_answers;`,
				createNewSlice: CreateAnswer,
				typ:            reflect.TypeOf(&Answer{}),
			},
		})
	return routeConfigs
}
