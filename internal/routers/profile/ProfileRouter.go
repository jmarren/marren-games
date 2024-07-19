package routers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jmarren/marren-games/internal/auth"
	"github.com/jmarren/marren-games/internal/awssdk"
	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/labstack/echo/v4"
)

func ProfileRouter(r *echo.Group) {
	r.GET("", getProfilePage)
	r.POST("/logout", func(c echo.Context) error {
		cookie := &http.Cookie{
			Name:     "auth",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-1 * time.Hour),
		}
		c.SetCookie(cookie)
		return controllers.RenderTemplate(c, "index", nil)
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

	return controllers.RenderTemplate(c, "profile", data)
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

func getProfilePage(c echo.Context) error {
	username := auth.GetFromClaims(auth.Username, c)

	data := struct {
		Username string
	}{
		Username: username.(string),
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
