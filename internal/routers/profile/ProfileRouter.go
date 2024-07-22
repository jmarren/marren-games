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

	uploadErr := awssdk.UploadToS3(f, username)
	if uploadErr != nil {
		fmt.Println("uploadError uploading to s3: ", uploadErr)
		return uploadErr
	}

	data := struct {
		Username string
	}{
		Username: username,
	}

	fmt.Println(data)

	return GetMyProfilePage(c)
}

func getProfilePhotoViewer(c echo.Context) error {
	username := auth.GetFromClaims(auth.Username, c)
	data := struct {
		Username string
	}{
		Username: username.(string),
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

	query := `
    SELECT (
      SELECT email
      FROM users
      WHERE id = :user_id
    ) AS email, 
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
    ) AS num_games,
    (
      SELECT SUM(score)
      FROM scores
      WHERE user_id = :user_id
    ) as total_points
    FROM (SELECT 1) AS dummy;
  `

	row := tx.QueryRowContext(ctx, query, userIdArg)

	var (
		emailRaw       sql.NullString
		numFriendsRaw  sql.NullInt64
		numGamesRaw    sql.NullInt64
		totalPointsRaw sql.NullInt64
	)

	err = row.Scan(&emailRaw, &numFriendsRaw, &numGamesRaw, &totalPointsRaw)
	if err != nil {
		tx.Rollback()
		return fail(err)
	}

	data := struct {
		Username    string
		Email       string
		NumFriends  int64
		NumGames    int64
		TotalPoints int64
	}{
		Username:    username.(string),
		Email:       emailRaw.String,
		NumFriends:  numFriendsRaw.Int64,
		NumGames:    numGamesRaw.Int64,
		TotalPoints: totalPointsRaw.Int64,
	}

	currentUrl := c.Request().Header.Get("Hx-Current-Url")
	if currentUrl[len(currentUrl)-7:] == "sign-in" {
		return controllers.RenderTemplate(c, "profile-after-login", data)
	}
	return controllers.RenderTemplate(c, "profile", data)
}

func getProfilePhotoUpload(c echo.Context) error {
	username := auth.GetFromClaims(auth.Username, c)

	data := struct {
		Username string
	}{
		Username: username.(string),
	}
	return controllers.RenderTemplate(c, "upload-profile-photo", data)
}
