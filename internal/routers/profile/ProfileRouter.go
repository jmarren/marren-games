package profile

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/awssdk"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func ProfileRouter(r *echo.Group) {
	r.GET("", GetMyProfilePage)
	r.POST("/logout", func(c echo.Context) error {
		cookie := &http.Cookie{
			Name:     "auth",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-1 * time.Hour),
		}
		c.SetCookie(cookie)
		return controllers.RenderTemplate(c, "index-after-logout", nil)
	})

	r.Use(CustomBodyLimit("2MB"))
	r.POST("/profile-photo", uploadProfilePhoto)
	r.GET("/ui/profile-photo-upload", getProfilePhotoUpload)
	r.GET("/ui/profile-photo", getProfilePhotoViewer)
}

func uploadProfilePhoto(c echo.Context) error {
	fmt.Println(c.Request().Header)

	f, err := c.FormFile("profileImage")
	if err != nil {
		fmt.Println("error during c.FormFile: ", err)
		return err
	}

	username := auth.GetFromClaims(auth.Username, c).(string)
	userId := auth.GetFromClaims(auth.UserId, c).(int)

	userIdArg := sql.Named("user_id", userId)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))
	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ProfileRouter, uploadProfilePhoto(), failed to start tx: %v", err)
	}
	defer tx.Rollback()

	query := `
    SELECT photo_version
    FROM users
    WHERE users.id = :user_id;
  `

	row := tx.QueryRowContext(ctx, query, userIdArg)

	var userPhotoVersionRaw sql.NullInt64

	err = row.Scan(&userPhotoVersionRaw)
	if err != nil {
		return fmt.Errorf("error getting photo version: %v", err)
	}

	newPhotoVersion := userPhotoVersionRaw.Int64 + 1

	uploadLocation := fmt.Sprintf("%v-%v", username, newPhotoVersion)

	uploadErr := awssdk.UploadToS3(f, uploadLocation)
	if uploadErr != nil {
		fmt.Println("uploadError uploading to s3: ", uploadErr)
		return uploadErr
	}

	newPhotoVersionArg := sql.Named("new_photo_version", newPhotoVersion)

	query = `
    UPDATE users
    SET photo_version = :new_photo_version,
        last_modified = CURRENT_TIMESTAMP
    WHERE id = :user_id;
  `

	_, err = tx.ExecContext(ctx, query, userIdArg, newPhotoVersionArg)
	if err != nil {
		return fmt.Errorf("profileRouter, uploadProfilePhoto(), error updating user photo version: %v", err)
	}
	tx.Commit()

	data := struct {
		Username string
	}{
		Username: username,
	}

	fmt.Println(data)

	c.Response().Header().Set("Hx-Push-Url", "/auth/profile")
	return GetMyProfilePage(c)
}

func getProfilePhotoViewer(c echo.Context) error {
	username := auth.GetFromClaims(auth.Username, c)
	userId := auth.GetFromClaims(auth.UserId, c)

	userIdArg := sql.Named("user_id", userId)

	query := `
    SELECT photo_version
    FROM users
    WHERE users.id = :user_id;
  `

	row := db.Sqlite.QueryRow(query, userIdArg)

	var userPhotoVersionRaw sql.NullInt64

	err := row.Scan(&userPhotoVersionRaw)
	if err != nil {
		return fmt.Errorf("error getting photo version: %v", err)
	}

	data := struct {
		Username     string
		PhotoVersion int64
	}{
		Username:     username.(string),
		PhotoVersion: userPhotoVersionRaw.Int64,
	}

	return controllers.RenderTemplate(c, "profile-photo-viewer", data)
}

func GetMyProfilePage(c echo.Context) error {
	fail := func(err error) error {
		return fmt.Errorf("error @ ProfileRouter, getProfilePage(): %v ", err)
	}

	username := auth.GetFromClaims(auth.Username, c)
	userId := auth.GetFromClaims(auth.UserId, c)

	// convert userId to sql.Named
	userIdArg := sql.Named("user_id", userId)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))
	defer cancel()

	tx, err := db.Sqlite.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	query := `
    SELECT last_modified
    FROM users
    WHERE users.id = :user_id;
  `

	row := tx.QueryRowContext(ctx, query, userIdArg)

	var lastModifiedStr string
	err = row.Scan(&lastModifiedStr)
	if err != nil {
		fail(fmt.Errorf(", scanning into lastModified var: %v", err))
	}
	var lastModified time.Time
	lastModified, err = time.Parse(time.RFC3339, lastModifiedStr)
	if err != nil {
		lastModified = time.Time{}
	}
	var ifModifiedSinceTime time.Time
	ifModifiedSince := c.Request().Header.Get(echo.HeaderIfModifiedSince)
	if ifModifiedSince != "" {
		ifModifiedSinceTime, err = time.Parse(http.TimeFormat, ifModifiedSince)
		if err != nil {
			ifModifiedSinceTime = time.Time{}
		}
	}
	if !ifModifiedSinceTime.IsZero() && lastModified.Before(ifModifiedSinceTime.Add(1*time.Second)) {
		tx.Commit()
		return c.NoContent(http.StatusNotModified)
	} else {
		c.Response().Header().Set(echo.HeaderCacheControl, "private, no-cache")
		c.Response().Header().Set(echo.HeaderLastModified, lastModified.Format(http.TimeFormat))
	}

	addTopRight := false
	addTopRightParam := c.QueryParam("add-top-right")
	if addTopRightParam == "true" {
		addTopRight = true
	}

	query = `
    SELECT email, photo_version,
    ( 
      SELECT COUNT(*)
      FROM friendships
  WHERE user_1_id = :user_id
  OR user_2_id = :user_id
    ) AS num_friends,
    (
      SELECT COUNT(*) 
      FROM user_game_membership
  WHERE user_id = :user_id
    ) AS num_games
        FROM users
  WHERE id = :user_id;
  `

	row = tx.QueryRowContext(ctx, query, userIdArg)

	var (
		emailRaw        sql.NullString
		photoVersionRaw sql.NullInt64
		numFriendsRaw   sql.NullInt64
		numGamesRaw     sql.NullInt64
	)

	err = row.Scan(&emailRaw, &photoVersionRaw, &numFriendsRaw, &numGamesRaw)
	if err != nil {
		return fail(err)
	}

	data := struct {
		Username     string
		Email        string
		PhotoVersion int64
		NumFriends   int64
		NumGames     int64
	}{
		Username:     username.(string),
		Email:        emailRaw.String,
		PhotoVersion: photoVersionRaw.Int64,
		NumFriends:   numFriendsRaw.Int64,
		NumGames:     numGamesRaw.Int64,
	}

	c.Response().Header().Set("Hx-Push-Url", "/auth/profile")

	tx.Commit()
	if addTopRight {
		return controllers.RenderTemplate(c, "profile-after-login", data)
	}
	return controllers.RenderTemplate(c, "profile", data)
}

func getProfilePhotoUpload(c echo.Context) error {
	username := auth.GetFromClaims(auth.Username, c)
	userId := auth.GetFromClaims(auth.UserId, c)

	userIdArg := sql.Named("user_id", userId)

	query := `
    SELECT photo_version
    FROM users
    WHERE users.id = :user_id;
  `

	row := db.Sqlite.QueryRow(query, userIdArg)

	var userPhotoVersionRaw sql.NullInt64

	err := row.Scan(&userPhotoVersionRaw)
	if err != nil {
		return fmt.Errorf("error getting photo version: %v", err)
	}

	data := struct {
		Username     string
		PhotoVersion int64
	}{
		Username:     username.(string),
		PhotoVersion: userPhotoVersionRaw.Int64,
	}
	return controllers.RenderTemplate(c, "upload-profile-photo", data)
}
