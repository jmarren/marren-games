package routers

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

func FriendsRouter(r *echo.Group) {
	r.GET("", getFriendsPage)

	r.POST("/search", searchUsers)
	r.GET("/profiles/:userId", getUserProfile)
	r.POST("/friend-requests/:user-id", createFriendRequest)
	r.POST("/friendships/:user-id", createFriendship)
	r.DELETE("/friend-requests/:user-id", deleteRequest)
}

func getFriendsPage(c echo.Context) error {
	myUserId := auth.GetFromClaims(auth.UserId, c)
	query := `
      SELECT fr.from_user_id, u.username
      FROM friend_requests fr
      JOIN users u ON fr.from_user_id = u.id
      WHERE to_user_id = ?;
  `
	rows, err := db.Sqlite.Query(query, myUserId)
	if err != nil {
		fmt.Println("error while querying for all friend requests: ", err)
		panic(err)
	}

	type FriendRequest struct {
		FromId       int
		FromUsername string
	}

	var friendRequests []FriendRequest

	for rows.Next() {
		var friendRequestId int
		var friendRequestUsername string
		rows.Scan(&friendRequestId, &friendRequestUsername)
		friendRequests = append(friendRequests,
			FriendRequest{
				FromId:       friendRequestId,
				FromUsername: friendRequestUsername,
			})
	}

	for _, friendRequest := range friendRequests {
		fmt.Println("friend request from: ", friendRequest.FromUsername)
	}

	data := struct {
		FriendRequests []FriendRequest
	}{
		FriendRequests: friendRequests,
	}

	return controllers.RenderTemplate(c, "friends", data)
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

func createFriendship(c echo.Context) error {
	myUserId := sql.NamedArg{
		Name:  "my_user_id",
		Value: auth.GetFromClaims(auth.UserId, c),
	}

	newFriendId := sql.NamedArg{
		Name:  "new_friend_id",
		Value: c.Param("user-id"),
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		cancel()
		return c.String(http.StatusInternalServerError, "error")
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()
	deleteRequestQuery := `
        DELETE FROM friend_requests
        WHERE from_user_id = :new_friend_id
        AND to_user_id = :my_user_id
          `
	_, err = db.Sqlite.ExecContext(ctx, deleteRequestQuery, myUserId, newFriendId)
	if err != nil {
		fmt.Println("error deleting friend request")
		return err
	}

	insertFriendshipQuery := `
        INSERT INTO friendships (user_1_id, user_2_id)
        VALUES (:my_user_id, :new_friend_id);`

	_, err = db.Sqlite.ExecContext(ctx, insertFriendshipQuery, myUserId, newFriendId)
	if err != nil {
		return err
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return err
	}

	return c.HTML(http.StatusOK, "Accepted!")
}

func deleteRequest(c echo.Context) error {
	myUserId := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))
	otherUserId := sql.Named("other_user_id", c.Param("user-id"))

	query := `
        DELETE FROM friend_requests
        WHERE (from_user_id = :other_user_id AND to_user_id = :my_user_id)
        OR
        (from_user_id = :my_user_id AND to_user_id = :other_user_id);
      `
	_, err := db.Sqlite.Exec(query, myUserId, otherUserId)
	if err != nil {
		fmt.Println("error deleting friend request: ", err)
		return err
	}
	return c.HTML(http.StatusOK, "Request Declined")
}
