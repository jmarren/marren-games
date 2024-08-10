package games

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func GameRouter(r *echo.Group) {
	r.GET("", getGameById)
	r.GET("/create-question", getCreateQuestionUI)
	r.GET("/invite-friends", getInviteFriendsUI)
	// TODO
	// r.PUT("", updateGame)
	// r.DELETE("", deleteGame)

	resultsGroup := r.Group("/results")
	// playGroup := r.Group("/play")
	questionsGroup := r.Group("/questions")
	playersGroup := r.Group("/players")
	invitesGroup := r.Group("/invites")

	ResultsRouter(resultsGroup)
	// PlayRouter(playGroup)
	QuestionsRouter(questionsGroup)
	PlayersRouter(playersGroup)
	InvitesRouter(invitesGroup)
}

func getGameById(c echo.Context) error {
	fail := func(err error) error {
		return fmt.Errorf("\033[31m getPlayPage error: %v \033[0m", err)
	}

	// Get Necessary Data from request
	gameId := c.Param("game-id")
	gameIdInt, err := strconv.Atoi(gameId)
	if err != nil {
		fmt.Println("game-id param is not convertible to int: ", err)
		return err
	}
	gameIdArg := sql.Named("game_id", gameIdInt)
	myUserId := auth.GetFromClaims(auth.UserId, c)
	myUserIdArg := sql.Named("my_user_id", myUserId)

	var ifModifiedSinceTime time.Time
	ifModifiedSince := c.Request().Header.Get(echo.HeaderIfModifiedSince)
	if ifModifiedSince != "" {
		ifModifiedSinceTime, err = time.Parse(http.TimeFormat, ifModifiedSince)
		if err != nil {
			ifModifiedSinceTime = time.Time{}
		}
	}

	// Create DB Transaction
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(8*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	var lastModifiedStr string

	// Query to get last modified timestamp
	query := `
    SELECT last_modified
    FROM games
    WHERE id = :game_id;
  `
	row := tx.QueryRowContext(ctx, query, gameIdArg)

	err = row.Scan(&lastModifiedStr)
	if err != nil {
		fmt.Println("error while scanning into vars")
		return err
	}

	var lastModified time.Time
	lastModified, err = time.Parse(time.RFC3339, lastModifiedStr)
	if err != nil {
		lastModified = time.Time{}
	}

	if !ifModifiedSinceTime.IsZero() && lastModified.Before(ifModifiedSinceTime.Add(1*time.Second)) {
		tx.Commit()
		return c.NoContent(http.StatusNotModified)
	} else {
		c.Response().Header().Set(echo.HeaderCacheControl, "private, no-cache")
		c.Response().Header().Set(echo.HeaderLastModified, lastModified.Format(http.TimeFormat))
	}

	query = `
    SELECT (
      CASE WHEN
    (
    SELECT user_id
    FROM current_askers
    WHERE current_askers.game_id = :game_id) = :my_user_id THEN 1
    ELSE 0
    END
    ) AS is_asker,
    (
      CASE WHEN (
        SELECT COUNT(*)
        FROM questions
        WHERE game_id = :game_id
          AND DATE(date_created, 'localtime') = DATE('now', 'localtime')
        ) > 0 THEN 1
      ELSE 0
      END)
    AS todays_question_created
    FROM current_askers;
  `
	var isAskerInt sql.NullInt64
	var todaysQuestionCreatedInt sql.NullInt64

	row = tx.QueryRowContext(ctx, query, gameIdArg, myUserIdArg)

	err = row.Scan(&isAskerInt, &todaysQuestionCreatedInt)
	if err != nil {
		fmt.Println("error while scanning into vars")
		return err
	}

	// convert variables to booleans
	var isAsker bool
	var todaysQuestionCreated bool

	if isAskerInt.Int64 == 1 {
		isAsker = true
	} else {
		isAsker = false
	}

	if todaysQuestionCreatedInt.Int64 == 1 {
		todaysQuestionCreated = true
	} else {
		todaysQuestionCreated = false
	}
	fmt.Printf("\n \033[31m isAsker: %v\ntodaysQuestionCreated: %v  \033[0m \n", isAsker, todaysQuestionCreated)

	// if the user is todays asker
	if isAsker {
		tx.Commit()
		if todaysQuestionCreated {
			return GetGameResults(c)
		} else {
			data := struct {
				GameId int
			}{
				GameId: gameIdInt,
			}
			return controllers.RenderTemplate(c, "create-question", data)
		}
	}

	// if todays question hasn't been created yet
	if !todaysQuestionCreated {
		return controllers.RenderTemplate(c, "no-question-yet", nil)
	}
	// Get todays question id
	questionId, err := GetTodaysQuestionId(gameIdInt)
	if err != nil {
		return err
	}
	questionIdArg := sql.Named("question_id", questionId)

	// determine if user has answered todays question
	query = `
    SELECT (
      CASE
      WHEN (
        SELECT COUNT(*)
        FROM answers
        WHERE answerer_id = :my_user_id
            AND game_id = :game_id
            AND question_id = :question_id
      ) > 0 THEN 1
      ELSE 0
      END
    ) AS result;
    `
	row = tx.QueryRowContext(ctx, query, myUserIdArg, gameIdArg, questionIdArg)

	// scan into var
	var answeredTodaysQuestionInt sql.NullInt64
	err = row.Scan(&answeredTodaysQuestionInt)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while querying to determine if user has answered todays question: %v", err)
	}

	var answeredTodaysQuestion bool

	if answeredTodaysQuestionInt.Int64 == 1 {
		answeredTodaysQuestion = true
	} else {
		answeredTodaysQuestion = false
	}

	// if they already answered todays question
	if answeredTodaysQuestion {

		fmt.Printf("\n \033[31m answeredTodaysQuestion: %v  \033[0m \n", answeredTodaysQuestion)

		tx.Commit()
		now := time.Now()
		c.Response().Header().Set("Expires", now.Add(time.Minute*15).Format(http.TimeFormat))
		// time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1).Format(http.TimeFormat))
		return GetGameResults(c)
	}

	query = `
    SELECT games.name, question_text, questions.id AS question_id, option_1, option_2, option_3, option_4
    FROM questions
    INNER JOIN games
      ON questions.game_id = games.id
    WHERE questions.game_id = :game_id AND DATE(questions.date_created, 'localtime') = DATE('now', 'localtime');
  `

	result := db.Sqlite.QueryRow(query, gameIdArg)

	if err != nil {
		fmt.Println("error querying for question: ", err)
		return err
	}

	var (
		questionRaw    sql.NullString
		questionIdRaw  sql.NullInt64
		gameNameRaw    sql.NullString
		optionOneRaw   sql.NullString
		optionTwoRaw   sql.NullString
		optionThreeRaw sql.NullString
		optionFourRaw  sql.NullString
	)

	err = result.Scan(&gameNameRaw, &questionRaw, &questionIdRaw, &optionOneRaw, &optionTwoRaw, &optionThreeRaw, &optionFourRaw)
	if err != nil {
		fmt.Println("error scanning question data into vars: ", err)
		return err
	}

	type GameData struct {
		GameId      int
		GameName    string
		Question    string
		QuestionId  int64
		OptionOne   string
		OptionTwo   string
		OptionThree string
		OptionFour  string
	}

	data := struct {
		ShowResults int
		GameData    GameData
	}{
		ShowResults: 0,
		GameData: GameData{
			GameId:      gameIdInt,
			GameName:    gameNameRaw.String,
			Question:    questionRaw.String,
			QuestionId:  questionIdRaw.Int64,
			OptionOne:   optionOneRaw.String,
			OptionTwo:   optionTwoRaw.String,
			OptionThree: optionThreeRaw.String,
			OptionFour:  optionFourRaw.String,
		},
	}

	tx.Commit()
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache, private")
	return controllers.RenderTemplate(c, "gameplay", data)
}

func getCreateQuestionUI(c echo.Context) error {
	// Content is static (only data is retrieved from url params) --> Set long cache policy
	lastModified := time.Date(2024, 7, 31, 0, 0, 0, 0, time.UTC)
	ifModifiedHeader := c.Request().Header.Get("If-Modified-Since")
	if ifModifiedHeader != "" {
		ifModifiedSinceTime, err := time.Parse(http.TimeFormat, ifModifiedHeader)
		if err != nil {
			return fmt.Errorf("if modified header not in proper format: %v ", err)
		}

		if !ifModifiedSinceTime.IsZero() && lastModified.Before(ifModifiedSinceTime.Add(1*time.Second)) {
			return c.NoContent(http.StatusNotModified)
		}
	}

	// Content is static (only data is retrieved from url params) --> Set long cache policy
	c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=15000")
	c.Response().Header().Set(echo.HeaderLastModified, lastModified.Format(http.TimeFormat))

	// Get Game Id
	gameId, err := strconv.Atoi(c.Param("game-id"))
	if err != nil {
		return fmt.Errorf("game-id param not convertible to int %v", err)
	}

	data := struct {
		GameId int
	}{
		GameId: gameId,
	}
	return controllers.RenderTemplate(c, "create-question", data)
}

func getInviteFriendsUI(c echo.Context) error {
	lastModified := time.Date(2024, 7, 31, 0, 0, 0, 0, time.UTC)
	ifModifiedHeader := c.Request().Header.Get("If-Modified-Since")
	if ifModifiedHeader != "" {
		ifModifiedSinceTime, err := time.Parse(http.TimeFormat, ifModifiedHeader)
		if err != nil {
			return fmt.Errorf("if modified header not in proper format: %v ", err)
		}

		if !ifModifiedSinceTime.IsZero() && lastModified.Before(ifModifiedSinceTime.Add(1*time.Second)) {
			return c.NoContent(http.StatusNotModified)
		}
	}

	// Content is static (only data is retrieved from url params) --> Set long cache policy
	c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=15000")
	c.Response().Header().Set(echo.HeaderLastModified, lastModified.Format(http.TimeFormat))
	gameId, err := strconv.Atoi(c.Param("game-id"))
	if err != nil {
		return fmt.Errorf("game-id param not convertible to int %v", err)
	}

	data := struct {
		GameId int
	}{
		GameId: gameId,
	}

	return controllers.RenderTemplate(c, "invite-friends", data)
}
