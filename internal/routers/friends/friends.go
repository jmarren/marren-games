package friends

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
	r.GET("/profiles/:user-id", getUserProfile)
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

	fromUrl := c.Request().Header.Get("Hx-Current-Url")
	shortenedUrl := fromUrl[len(fromUrl)-14:]

	fmt.Println("()()()() shortededUrl: ", shortenedUrl)

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

	if shortenedUrl != "invite-friends" {
		type DataStruct struct {
			Data []struct {
				Username string
				Email    string
				UserId   int64
			}
			GameId int
		}

		dataStruct := DataStruct{
			Data:   users,
			GameId: -1,
		}
		return controllers.RenderTemplate(c, "search-results", dataStruct)
	}

	gameId := c.QueryParam("game-id")
	gameIdInt, err := strconv.Atoi(gameId)
	if err != nil {
		fmt.Println("error: game-id from params not convertible to int")
		return err
	}
	type DataStruct struct {
		Data []struct {
			Username string
			Email    string
			UserId   int64
		}
		GameId int
	}

	dataStruct := DataStruct{
		Data:   users,
		GameId: gameIdInt,
	}

	return controllers.RenderTemplate(c, "search-results", dataStruct)
}

func getUserProfile(c echo.Context) error {
	otherUserId, err := strconv.Atoi(c.Param("user-id"))
	if err != nil {
		return fmt.Errorf("friends, getUserProfile(): %v", err)
	}
	myUserId, ok := auth.GetFromClaims(auth.UserId, c).(int)
	if !ok {
		return fmt.Errorf("myUserId from claims failed assertion to int:  %v", nil)
	}
	otherUserIdArg := sql.Named("other_user_id", c.Param("user-id"))
	myUserIdArg := sql.Named("my_user_id", myUserId)

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
    FROM friendships) AS is_friend,
    ( 
      SELECT COUNT(*)
      FROM friendships
      WHERE user_1_id = :other_user_id
          OR user_2_id = :other_user_id
    ) AS num_friends,
    (
      SELECT COUNT(*) 
      FROM user_game_membership
      WHERE user_id = :other_user_id
    ) AS num_games,
    (
      SELECT SUM(score)
      FROM scores
      WHERE user_id = :other_user_id
    ) as total_points
    FROM users
  WHERE id = :other_user_id;
  `
	row := db.Sqlite.QueryRow(query, myUserIdArg, otherUserIdArg)

	var (
		usernameRaw    sql.NullString
		emailRaw       sql.NullString
		requestedRaw   sql.NullInt64
		isFriendRaw    sql.NullInt64
		numFriendsRaw  sql.NullInt64
		numGamesRaw    sql.NullInt64
		totalPointsRaw sql.NullInt64
	)
	err = row.Scan(&usernameRaw, &emailRaw, &requestedRaw, &isFriendRaw, &numFriendsRaw, &numGamesRaw, &totalPointsRaw)
	if err != nil {
		fmt.Println("error querying db:", err)
		panic(err)
	}

	data := struct {
		Username    string
		UserId      int
		Email       string
		Requested   int64
		IsFriend    int64
		NumFriends  int64
		NumGames    int64
		TotalPoints int64
	}{
		Username:    usernameRaw.String,
		UserId:      otherUserId,
		Email:       emailRaw.String,
		Requested:   requestedRaw.Int64,
		IsFriend:    isFriendRaw.Int64,
		NumFriends:  numFriendsRaw.Int64,
		NumGames:    numGamesRaw.Int64,
		TotalPoints: totalPointsRaw.Int64,
	}

	c.Response().Header().Set("Hx-Push-Url", fmt.Sprintf("/auth/friends/profiles/%v", otherUserId))
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
