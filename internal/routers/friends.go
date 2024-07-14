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
	r.DELETE("/friendships/:user-id", deleteFriendship)
}

func getFriendsPage(c echo.Context) error {
	myUserId := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))
	query := `
      SELECT fr.from_user_id, u.username
      FROM friend_requests fr
      JOIN users u ON fr.from_user_id = u.id
  WHERE to_user_id = :my_user_id;
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
		err := rows.Scan(&friendRequestId, &friendRequestUsername)
		if err != nil {
			return err
		}
		friendRequests = append(friendRequests,
			FriendRequest{
				FromId:       friendRequestId,
				FromUsername: friendRequestUsername,
			})
	}

	for _, friendRequest := range friendRequests {
		fmt.Println("friend request from: ", friendRequest.FromUsername)
	}

	query = `
      SELECT
        CASE
  WHEN friendships.user_1_id = :my_user_id
            THEN friendships.user_2_id
  WHEN friendships.user_2_id = :my_user_id
            THEN friendships.user_1_id
        END AS friend_id,
        users.username AS friend_username
        FROM friendships
        JOIN users ON users.id = friend_id;
  `

	rows, err = db.Sqlite.Query(query, myUserId)
	if err != nil {
		fmt.Println("error while querying for all friend requests: ", err)
		panic(err)
	}

	type Friend struct {
		Username string
		UserId   int
	}
	var friends []Friend
	for rows.Next() {
		var friend Friend
		err := rows.Scan(&friend.UserId, &friend.Username)
		if err != nil {
			fmt.Println("error querying db for friends: ", err)
			return err
		}
		friends = append(friends, friend)
	}

	data := struct {
		FriendRequests []FriendRequest
		Friends        []Friend
	}{
		FriendRequests: friendRequests,
		Friends:        friends,
	}

	return controllers.RenderTemplate(c, "friends", data)
}

func searchUsers(c echo.Context) error {
	searchParam := c.FormValue("search")
	fmt.Println("()()()() \nquery received for: ", searchParam)
	if searchParam == "" {
		return c.HTML(http.StatusOK, "")
	}
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
	otherUserId := sql.Named("other_user_id", c.Param("userId"))
	myUserId := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))

	query := `
    SELECT username, email,
    (SELECT MAX(
        CASE
  WHEN from_user_id = :my_user_id AND to_user_id = :other_user_id THEN 1
          ELSE 0
      END)
    FROM friend_requests) AS requested,
    (SELECT MAX(
      CASE
      WHEN (friendships.user_1_id = :my_user_id AND friendships.user_2_id = :other_user_id)
      OR
      (friendships.user_1_id = :other_user_id AND friendships.user_2_id = :my_user_id)
      THEN 1
      ELSE 0
    END)
    FROM friendships) AS is_friend
    FROM users
  WHERE id = :other_user_id;
  `
	row := db.Sqlite.QueryRow(query, myUserId, otherUserId)

	var (
		username  string
		email     string
		requested sql.NullInt64
		isFriend  sql.NullInt64
	)
	err := row.Scan(&username, &email, &requested, &isFriend)
	if err != nil {
		fmt.Println("error querying db:", err)
		panic(err)
	}

	otherUserIdInt, err := strconv.Atoi(otherUserId.Value.(string))
	if err != nil {
		fmt.Println("userId is not convertible to int", err)
		panic(err)
	}

	var requestedVal int
	var isFriendVal int
	if requested.Int64 == 0 || !requested.Valid {
		requestedVal = 0
	} else {
		requestedVal = 1
	}

	if isFriend.Int64 == 0 || !isFriend.Valid {
		isFriendVal = 0
	} else {
		isFriendVal = 1
	}

	data := struct {
		Username  string
		UserId    int
		Email     string
		Requested int
		IsFriend  int
	}{
		Username:  username,
		UserId:    otherUserIdInt,
		Email:     email,
		Requested: requestedVal,
		IsFriend:  isFriendVal,
	}

	return controllers.RenderTemplate(c, "other-user-profile", data)
}

func createFriendRequest(c echo.Context) error {
	myUserId := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))
	otherUserId := sql.Named("other_user_id", c.Param("user-id"))
	otherUserIdInt, err := strconv.Atoi(c.Param("user-id"))
	if err != nil {
		fmt.Println("error converting toUserId to int")
		return err
	}

	query := `
      INSERT INTO friend_requests (from_user_id, to_user_id)
  VALUES(:my_user_id, :other_user_id);
  `

	_, err = db.Sqlite.Exec(query, myUserId, otherUserId)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if err != nil {
		fmt.Println("error: otherUserId is not convertible to int: ", err)
		return err
	}

	data := struct {
		UserId int
	}{
		UserId: otherUserIdInt,
	}
	return controllers.RenderTemplate(c, "request-sent-button", data)
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
	otherId, err := strconv.Atoi(c.Param("user-id"))
	if err != nil {
		fmt.Println("error converting user-id url param to int: ", err)
		return err
	}
	otherUserId := sql.Named("other_user_id", otherId)
	query := `
          SELECT MAX(
            CASE 
                WHEN (from_user_id = :my_user_id AND to_user_id = :other_user_id) THEN 1
                ELSE 0
              END) AS is_requester
          FROM friend_requests;
  `

	result := db.Sqlite.QueryRow(query, myUserId, otherUserId)

	var isRequester sql.NullInt64
	err = result.Scan(&isRequester)
	if err != nil {
		fmt.Println("error while querying to determine who requested:", err)
		return err
	}

	fmt.Println()
	fmt.Println("@@@@ isRequester: ", isRequester)
	fmt.Println()
	var isRequesterVal int
	if !isRequester.Valid || isRequester.Int64 == 0 {
		isRequesterVal = 0
	} else {
		isRequesterVal = 1
	}
	fmt.Println()
	fmt.Println("@@@@ isRequesterVal: ", isRequesterVal)
	fmt.Println()
	query = `
        DELETE FROM friend_requests
        WHERE (from_user_id = :other_user_id AND to_user_id = :my_user_id)
        OR
        (from_user_id = :my_user_id AND to_user_id = :other_user_id);
      `
	_, err = db.Sqlite.Exec(query, myUserId, otherUserId)
	if err != nil {
		fmt.Println("error deleting friend request: ", err)
		return err
	}

	if isRequesterVal == 1 {
		data := struct {
			UserId int
		}{
			UserId: otherId,
		}
		return controllers.RenderTemplate(c, "add-friend-button", data)
	}
	return c.HTML(http.StatusOK, "Request Declined")
}

func deleteFriendship(c echo.Context) error {
	myUserId := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))
	otherId, err := strconv.Atoi(c.Param("user-id"))
	if err != nil {
		fmt.Println("error converting user-id url param to int: ", err)
		return err
	}
	otherUserId := sql.Named("other_user_id", otherId)

	query := `
  DELETE FROM friendships
  WHERE (friendships.user_1_id = :my_user_id AND friendships.user_2_id = :other_user_id) 
      OR
        (friendships.user_1_id = :other_user_id AND friendships.user_2_id = :my_user_id);
`
	_, err = db.Sqlite.Exec(query, myUserId, otherUserId)
	if err != nil {
		fmt.Println("error deleting friendship: ", err)
		return err
	}

	data := struct {
		UserId int
	}{
		UserId: otherId,
	}

	return controllers.RenderTemplate(c, "add-friend-button", data)
}
