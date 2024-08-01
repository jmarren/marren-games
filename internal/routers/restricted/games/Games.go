package games

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func GamesRouter(r *echo.Group) {
	r.GET("", getGames)
	r.POST("", createGame)
	r.GET("/create", getCreateGameUI)

	gameGroup := r.Group("/:game-id")
	GameRouter(gameGroup)
}

func getGames(c echo.Context) error {
	fmt.Println("hit get games")
	// Set Header to revalidate cache
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache, private")
	// Get User Id
	myUserId := auth.GetFromClaims(auth.UserId, c)

	// Get All Games where User is a Member
	// create named arg
	myUserIdArg := sql.Named("my_user_id", myUserId)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	// begin the Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return fmt.Errorf("getGames(), error beginning db tx: %v", err)
	}

	defer tx.Rollback()

	// query to get most recent last_modified for all the games the user is a member of
	query := `
WITH games_last_modified AS (
	SELECT 1 AS id, last_modified AS last_date
	FROM games
	JOIN user_game_membership
		ON user_game_membership.game_id = games.id
  WHERE user_game_membership.user_id = :my_user_id
	ORDER BY games.last_modified DESC
	LIMIT 1),
	most_recent_invite AS (
	SELECT 1 AS id,  date_invited   AS last_date
	FROM user_game_invites
  WHERE user_game_invites.user_id = :my_user_id
	ORDER BY date_invited DESC
	LIMIT 1)
	SELECT MAX(IFNULL(games_last_modified.last_date, 0), IFNULL(most_recent_invite.last_date, 0)) AS most_recent_change
	FROM games_last_modified
	LEFT JOIN most_recent_invite
		ON games_last_modified.id = most_recent_invite.id;
  `

	var ifModifiedSinceTime time.Time
	ifModifiedSince := c.Request().Header.Get(echo.HeaderIfModifiedSince)
	if ifModifiedSince != "" {
		ifModifiedSinceTime, err = time.Parse(http.TimeFormat, ifModifiedSince)
		fmt.Printf("\n ifModifiedSinceTime: %v", ifModifiedSinceTime)
		if err != nil {
			fmt.Printf("\nerror: no if-modified-since header: %v", err)
			ifModifiedSinceTime = time.Time{}
		}
	}

	row := tx.QueryRowContext(ctx, query, myUserIdArg)

	var lastModifiedStr string

	err = row.Scan(&lastModifiedStr)
	if err != nil {
		return fmt.Errorf("/games getGames(), scanning into lastModified: %v", err)
	}

	fmt.Printf("\n lastModifiedStr: %v", lastModifiedStr)
	fmt.Printf("\n ifModifiedSinceTime: %v", ifModifiedSinceTime)

	var lastModified time.Time
	lastModified, err = time.Parse(time.DateTime, lastModifiedStr)
	if err != nil {
		fmt.Printf("\nPOTENTIAL ERROR: Games, getGames(), error while parsing lastModifiedStr %v ", err)
		lastModified = time.Time{}
	}

	if !ifModifiedSinceTime.IsZero() && lastModified.Before(ifModifiedSinceTime.Add(1*time.Second)) {
		tx.Commit()
		return c.NoContent(http.StatusNotModified)
	} else {
		fmt.Println("if-modified-since is 0 or before last_modified")
		fmt.Println(" setting lastModified header to: ", lastModified.Format(http.TimeFormat))
		c.Response().Header().Set(echo.HeaderCacheControl, "private, no-cache")
		c.Response().Header().Set(echo.HeaderLastModified, lastModified.Format(http.TimeFormat))
	}

	type Game struct {
		GameId               int64
		GameName             string
		GameTotalMembers     int64
		QuestionText         string
		CurrentAskerUsername string
	}

	var games []Game

	query = `
SELECT user_game_membership.game_id, games.name, member_counts.total_members, question_text,  users.username AS current_asker_username
  FROM user_game_membership
JOIN games
 	ON games.id = user_game_membership.game_id
JOIN current_askers
	ON current_askers.game_id = games.id
JOIN users
	ON users.id = current_askers.user_id
LEFT JOIN questions ON
	user_game_membership.game_id = questions.game_id
	  AND DATE(questions.date_created) = DATE('now') 
JOIN (
      SELECT game_id , COUNT(user_id) AS total_members
  FROM user_game_membership
  GROUP BY game_id) AS member_counts
    ON member_counts.game_id = user_game_membership.game_id
  WHERE user_game_membership.user_id = :my_user_id;
  `

	rows, err := tx.QueryContext(ctx, query, myUserIdArg)
	if err != nil {
		fmt.Println("error querying db for user's game ids:", err)
		return err
	}

	for rows.Next() {
		var gameIdRaw sql.NullInt64
		var gameNameRaw sql.NullString
		var totalMembers sql.NullInt64
		var questionTextRaw sql.NullString
		var currentAskerUsernameRaw sql.NullString

		err := rows.Scan(&gameIdRaw, &gameNameRaw, &totalMembers, &questionTextRaw, &currentAskerUsernameRaw)
		if err != nil {
			fmt.Println("error scanning game_id into gameId variable: ", err)
			return err
		}
		if !gameIdRaw.Valid || !gameNameRaw.Valid {
			fmt.Println("gameId not valid:", gameIdRaw, gameNameRaw)
		}

		games = append(games, Game{
			GameId:               gameIdRaw.Int64,
			GameName:             gameNameRaw.String,
			GameTotalMembers:     totalMembers.Int64,
			QuestionText:         questionTextRaw.String,
			CurrentAskerUsername: currentAskerUsernameRaw.String,
		})
	}

	fmt.Println("%%%%%%%%%% games: ", games, "%%%%%%%%%%%%")

	// // Get All Invites for the User
	query = `
	       SELECT name, creator_id, game_id, users.username AS creator_name
         FROM user_game_invites
         LEFT JOIN games
          ON games.id = game_id
         LEFT JOIN users
          ON creator_id = users.id
         WHERE user_id = :my_user_id;
	 `

	var gameInvites []GameInvite

	rows, err = tx.QueryContext(ctx, query, myUserIdArg)
	if err != nil {
		tx.Rollback()
		fmt.Println("error querying for game invites: ", err)
		return err
	}

	for rows.Next() {
		var (
			gameNameRaw    sql.NullString
			creatorIdRaw   sql.NullInt64
			gameIdRaw      sql.NullInt64
			creatorNameRaw sql.NullString
		)

		err := rows.Scan(&gameNameRaw, &creatorIdRaw, &gameIdRaw, &creatorNameRaw)
		if err != nil {
			fmt.Println("error scanning invites into vars: ", err)
			return err
		}
		gameInvites = append(gameInvites, GameInvite{
			GameName:    gameNameRaw.String,
			CreatorId:   creatorIdRaw.Int64,
			GameId:      gameIdRaw.Int64,
			CreatorName: creatorNameRaw.String,
		})

	}
	fmt.Println("%%%%%%%%%% Game Invites: ", gameInvites, "  %%%%%%%%%%%%%")

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("games, getGames(), error commiting tx: %v", err)
	}
	data := struct {
		CurrentGames []Game
		GameInvites  []GameInvite
	}{
		CurrentGames: games,
		GameInvites:  gameInvites,
	}
	return controllers.RenderTemplate(c, "games", data)
}

func getCreateGameUI(c echo.Context) error {
	// Content is static --> Set long cache policy
	c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=15000")
	return controllers.RenderTemplate(c, "create-game", nil)
}

func createGame(c echo.Context) error {
	userId := auth.GetFromClaims("UserId", c)
	gameName := c.FormValue("game-name")

	myUserIdArg := sql.Named("my_user_id", userId)
	gameNameArg := sql.Named("my_game_name", gameName)

	if gameName == "" {
		fmt.Println(" NAME NOT PROVIDED")
		return c.HTML(http.StatusBadRequest, "please provide a name")
	}

	// create context for query
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	// begin the Tx
	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}

	defer tx.Rollback()

	// Insert new game into games table
	result, err := tx.ExecContext(ctx, `
    INSERT INTO games (creator_id, name) VALUES (:my_user_id, :my_game_name);
    `, myUserIdArg, gameNameArg)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error: Games.go: inserting into games, %v", err)
	}
	gameId, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return errors.New("an error occurred")
	}

	data := struct {
		GameId int64
	}{
		GameId: gameId,
	}

	tx.Commit()
	c.Response().Header().Set("Hx-Push-Url", fmt.Sprintf("/auth/games/%v/invite-friends", gameId))
	return controllers.RenderTemplate(c, "invite-friends", data)
}
