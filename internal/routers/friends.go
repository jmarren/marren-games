package routers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func FriendsRouter(r *echo.Group) {
	r.GET("", getFriendsPage)

	r.POST("/search", searchUsers)
	r.GET("/profiles/:userId", getUserProfile)
	r.POST("/friend-requests/:user-id", createFriendRequest)
}

func getFriendsPage(c echo.Context) error {
	return controllers.RenderTemplate(c, "friends", nil)
}

func searchUsers(c echo.Context) error {
	searchParam := c.FormValue("search")
	fmt.Println("()()()() \nquery received for: ", searchParam)

	query := `
    SELECT username, email, id
    FROM users
    WHERE username LIKE ?;
  `

	rows, err := db.Sqlite.Query(query, searchParam+"%")
	if err != nil {
		fmt.Println("error querying db:", err)
		return err
	}

	var users []struct {
		Username string
		Email    string
		UserId   int64
	}

	for rows.Next() {
		var (
			username sql.NullString
			email    sql.NullString
			userId   sql.NullInt64
		)
		if err := rows.Scan(&username, &email, &userId); err != nil {
			fmt.Println("error scanning rows:", err)
			return err
		}

		fmt.Println(" \n\n username: ", username.String)
		users = append(users, struct {
			Username string
			Email    string
			UserId   int64
		}{
			Username: username.String,
			Email:    email.String,
			UserId:   userId.Int64,
		})
	}

	for _, user := range users {
		fmt.Println(user.Username)
		fmt.Println(user.Email)
	}

	type DataStruct struct {
		Data []struct {
			Username string
			Email    string
			UserId   int64
		}
	}

	dataStruct := DataStruct{
		Data: users,
	}

	return controllers.RenderTemplate(c, "search-results", dataStruct)
}

func getUserProfile(c echo.Context) error {
	profileUserId := c.Param("userId")
	fmt.Println(profileUserId)
	userId := auth.GetFromClaims(auth.UserId, c)

	query := `
    SELECT username, email,
    (SELECT MAX(
        CASE
        WHEN from_user_id = ? AND to_user_id = ? THEN 1
          ELSE 0
      END)
    FROM friend_requests) AS requested
    FROM users
    WHERE id = ?
  `
	row := db.Sqlite.QueryRow(query, userId, profileUserId, profileUserId)

	var (
		username  string
		email     string
		requested int64
	)
	err := row.Scan(&username, &email, &requested)
	if err != nil {
		fmt.Println("error querying db:", err)
		panic(err)
	}

	profileUserIdInt, err := strconv.Atoi(profileUserId)
	if err != nil {
		fmt.Println("userId is not convertible to int", err)
		panic(err)
	}

	data := struct {
		Username  string
		UserId    int
		Email     string
		Requested int64
	}{
		Username:  username,
		UserId:    profileUserIdInt,
		Email:     email,
		Requested: requested,
	}

	return controllers.RenderTemplate(c, "other-user-profile", data)
}

func createFriendRequest(c echo.Context) error {
	from_user_id := auth.GetFromClaims(auth.UserId, c)
	to_user_id := c.Param("user-id")
	query := `
      INSERT INTO friend_requests (from_user_id, to_user_id)
      VALUES(?, ?);
  `

	_, err := db.Sqlite.Exec(query, from_user_id, to_user_id)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return c.HTML(http.StatusOK, `
        <style>
          #add-friend-button {
            background-color: skyblue;
            color: navy;
          }
        </style>
        Requested
    `)
}
