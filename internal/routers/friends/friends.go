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
	myUserIdArg := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error begging tx, getFriendsPage(): %v", err)
	}
	defer tx.Rollback()

	// determine if friends or friend requests have changed
	query := `
  WITH dummy AS (
		SELECT 1 AS id
	),
	most_recent_friendship AS (
      SELECT 1 AS id, date_created AS last_date_created
      FROM friendships
      WHERE user_1_id = :my_user_id
         OR user_2_id = :my_user_id
      ORDER BY date_created DESC
      LIMIT 1
    ),
    most_recent_friend_request AS (
      SELECT 1 AS id, date_sent AS last_date_sent
      FROM friend_requests
      WHERE to_user_id = :my_user_id
      ORDER BY date_sent DESC
      LIMIT 1
    )
	SELECT
	MAX(
		IFNULL(most_recent_friendship.last_date_created, 0),
		IFNULL(most_recent_friend_request.last_date_sent, 0)
		)
	FROM dummy
	LEFT JOIN  most_recent_friendship
		ON  most_recent_friendship.id = dummy.id
	LEFT JOIN most_recent_friend_request
		ON most_recent_friend_request.id = dummy.id;
  `

	row := tx.QueryRowContext(ctx, query, myUserIdArg)

	var lastModifiedStr string

	err = row.Scan(&lastModifiedStr)
	if err != nil {
		return fmt.Errorf("friends, getFriendsPage(), error scanning into lastModifiedStr: %v", err)
	}

	var lastModified time.Time
	lastModified, err = time.Parse(time.DateTime, lastModifiedStr)
	if err != nil {
		fmt.Printf("\nPOTENTIAL ERROR: Games, getGames(), error while parsing lastModifiedStr %v \n ", err)
		lastModified = time.Time{}
	}

	ifModifiedSinceHeader := c.Request().Header.Get(echo.HeaderIfModifiedSince)

	ifModifiedSinceTime := time.Time{}
	if ifModifiedSinceHeader != "" {
		t, err := time.Parse(http.TimeFormat, ifModifiedSinceHeader)
		if err != nil {
			return fmt.Errorf("if-modified-since header not in proper format: %v", err)
		}
		ifModifiedSinceTime = t
		fmt.Println(ifModifiedSinceTime)
	}

	if !ifModifiedSinceTime.IsZero() && lastModified.Before(ifModifiedSinceTime.Add(1*time.Second)) {
		tx.Commit()
		return c.NoContent(http.StatusNotModified)
	} else {

		c.Response().Header().Set(echo.HeaderCacheControl, "private, no-cache")
		c.Response().Header().Set(echo.HeaderLastModified, lastModified.Format(http.TimeFormat))
	}

	query = `
      SELECT fr.from_user_id, u.username
      FROM friend_requests fr
      JOIN users u ON fr.from_user_id = u.id
      WHERE to_user_id = :my_user_id;
  `
	rows, err := tx.QueryContext(ctx, query, myUserIdArg)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("friends, getFriendsPage(), error querying for friend requests: %v", err)
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
			tx.Rollback()
			return err
		}
		friendRequests = append(friendRequests,
			FriendRequest{
				FromId:       friendRequestId,
				FromUsername: friendRequestUsername,
			})
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

	rows, err = tx.QueryContext(ctx, query, myUserIdArg)
	if err != nil {
		tx.Rollback()
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
			return fmt.Errorf("friends, getFriendsPage() error querying for friends: %v ", err)
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

	tx.Commit()

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

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("friends, searchUsers(), error beginning tx:  %v", err)
	}
	defer tx.Rollback()

	query := `
    SELECT username, email, id
    FROM users
    WHERE username LIKE ?
    LIMIT 10;
  `

	rows, err := tx.QueryContext(ctx, query, searchParam+"%")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("friends, searchUsers(), error querying for users: %v", err)
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
			tx.Rollback()
			return fmt.Errorf("friends, searchUsers(), error scanning into users struct: %v", err)
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
		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("friends, searchUsers(), error commiting tx: %v", err)
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
	// Get necessary data
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

	// begin db transaction
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("friends, getUserProfile(), error beginning tx:  %v", err)
	}
	defer tx.Rollback()

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
    ) AS num_games
       FROM users
  WHERE id = :other_user_id;
  `
	row := tx.QueryRowContext(ctx, query, myUserIdArg, otherUserIdArg)

	var (
		usernameRaw   sql.NullString
		emailRaw      sql.NullString
		requestedRaw  sql.NullInt64
		isFriendRaw   sql.NullInt64
		numFriendsRaw sql.NullInt64
		numGamesRaw   sql.NullInt64
	)
	err = row.Scan(&usernameRaw, &emailRaw, &requestedRaw, &isFriendRaw, &numFriendsRaw, &numGamesRaw)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("friends, getUserProfile(), error scanning into vars: %v", err)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("friends, getUserProfile(), error commiting tx: %v", err)
	}

	data := struct {
		Username   string
		UserId     int
		Email      string
		Requested  int64
		IsFriend   int64
		NumFriends int64
		NumGames   int64
	}{
		Username:   usernameRaw.String,
		UserId:     otherUserId,
		Email:      emailRaw.String,
		Requested:  requestedRaw.Int64,
		IsFriend:   isFriendRaw.Int64,
		NumFriends: numFriendsRaw.Int64,
		NumGames:   numGamesRaw.Int64,
	}
	c.Response().Header().Set("Hx-Push-Url", fmt.Sprintf("/auth/friends/profiles/%v", otherUserId))
	return controllers.RenderTemplate(c, "other-user-profile", data)
}

func createFriendRequest(c echo.Context) error {
	// Get necessary data
	myUserId := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))
	otherUserId := sql.Named("other_user_id", c.Param("user-id"))
	otherUserIdInt, err := strconv.Atoi(c.Param("user-id"))
	if err != nil {
		fmt.Println("error converting toUserId to int")
		return err
	}

	// begin db transaction
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("friends, createFriendRequest(), error beginning tx:  %v", err)
	}
	defer tx.Rollback()

	query := `
      INSERT INTO friend_requests (from_user_id, to_user_id)
  VALUES(:my_user_id, :other_user_id);
  `

	_, err = tx.ExecContext(ctx, query, myUserId, otherUserId)
	if err != nil {
		return fmt.Errorf("friends, createFriendRequest(), error inserting into friends requests: %v", err)
	}

	tx.Commit()

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
		return fmt.Errorf("friends, createFriendship(), error beginning tx: %v", err)
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
		return fmt.Errorf("friends, createFriendship(), error deleting friend request: %v", err)
	}

	insertFriendshipQuery := `
        INSERT INTO friendships (user_1_id, user_2_id)
        VALUES (:my_user_id, :new_friend_id);`

	_, err = db.Sqlite.ExecContext(ctx, insertFriendshipQuery, myUserId, newFriendId)
	if err != nil {
		return fmt.Errorf("friends, createFriendship(), error whil inserting into freindships: %v", err)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("friends, createFriendship(), error commiting tx: %v", err)
	}
	return c.HTML(http.StatusOK, "Accepted!")
}

func deleteRequest(c echo.Context) error {
	// get necessary data
	myUserId := sql.Named("my_user_id", auth.GetFromClaims(auth.UserId, c))
	otherId, err := strconv.Atoi(c.Param("user-id"))
	if err != nil {
		fmt.Println("error converting user-id url param to int: ", err)
		return err
	}
	otherUserId := sql.Named("other_user_id", otherId)

	// begin db trasnaction
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("friends, deleteRequest(), error beginning tx: %v", err)
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	query := `
          SELECT MAX(
            CASE 
                WHEN (from_user_id = :my_user_id AND to_user_id = :other_user_id) THEN 1
                ELSE 0
              END) AS is_requester
          FROM friend_requests;
  `

	result := tx.QueryRowContext(ctx, query, myUserId, otherUserId)

	var isRequester sql.NullInt64
	err = result.Scan(&isRequester)
	if err != nil {
		return fmt.Errorf("friends, deleteRequest(), error querying for isRequester: %v", err)
	}

	var isRequesterVal int
	if !isRequester.Valid || isRequester.Int64 == 0 {
		isRequesterVal = 0
	} else {
		isRequesterVal = 1
	}

	query = `
        DELETE FROM friend_requests
        WHERE (from_user_id = :other_user_id AND to_user_id = :my_user_id)
        OR
        (from_user_id = :my_user_id AND to_user_id = :other_user_id);
      `
	_, err = tx.ExecContext(ctx, query, myUserId, otherUserId)
	if err != nil {
		return fmt.Errorf("friends, deleteRequest(), error deleting requst: %v", err)
	}

	if isRequesterVal == 1 {
		data := struct {
			UserId int
		}{
			UserId: otherId,
		}
		return controllers.RenderTemplate(c, "add-friend-button", data)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("friends, deleteRequest(), error commiting tx: %v", err)
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

	// begin db trasnaction
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))

	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("friends, deleteFriendship(), error beginning tx: %v", err)
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	query := `
  DELETE FROM friendships
  WHERE (friendships.user_1_id = :my_user_id AND friendships.user_2_id = :other_user_id) 
      OR
        (friendships.user_1_id = :other_user_id AND friendships.user_2_id = :my_user_id);
`
	_, err = tx.ExecContext(ctx, query, myUserId, otherUserId)
	if err != nil {
		return fmt.Errorf("friends, deleteFriendship(), error deleting friendship: %v", err)
	}

	data := struct {
		UserId int
	}{
		UserId: otherId,
	}
	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("friends, deleteFriendship(), error commiting tx: %v", err)
	}
	return controllers.RenderTemplate(c, "add-friend-button", data)
}
