package games

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func QuestionsRouter(r *echo.Group) {
	r.POST("", createQuestion)
	r.GET("", getTodaysQuestion)
}

func createQuestion(c echo.Context) error {
	question := c.FormValue("question")
	optionOne := c.FormValue("option-1")
	optionTwo := c.FormValue("option-2")
	optionThree := c.FormValue("option-3")
	optionFour := c.FormValue("option-4")
	gameId := c.QueryParam("game-id")
	askerId := auth.GetFromClaims(auth.UserId, c)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error creating tx QuestionsRouter, createQuestion(): %v", err)
	}
	defer tx.Rollback()
	query := `
  INSERT OR IGNORE INTO questions (game_id,asker_id, question_text, option_1, option_2, option_3, option_4)
  VALUES (?, ?, ?, ?, ?, ?, ?);
  `

	_, err = tx.ExecContext(ctx, query, gameId, askerId, question, optionOne, optionTwo, optionThree, optionFour)
	if err != nil {
		tx.Rollback()
		return c.HTML(http.StatusBadRequest, err.Error())
	}

	data := struct {
		AnswersData    []AnswerStats
		ScoreboardData []UserScore
	}{
		AnswersData:    []AnswerStats{},
		ScoreboardData: []UserScore{},
	}

	if err := tx.Commit(); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	c.Response().Header().Set("Hx-Push-Url", fmt.Sprintf("/auth/games/%v/results", gameId))
	return controllers.RenderTemplate(c, "results", data)
}

func getTodaysQuestion(c echo.Context) error {
	return nil
}
