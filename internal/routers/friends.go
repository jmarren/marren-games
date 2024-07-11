package routers

import (
	"database/sql"
	"fmt"

	"github.com/jmarren/marren-games/internal/controllers"
	"github.com/jmarren/marren-games/internal/db"
	"github.com/labstack/echo/v4"
)

func FriendsRouter(r *echo.Group) {
	r.GET("", getFriendsPage)

	r.POST("/search", searchUsers)
}

func getFriendsPage(c echo.Context) error {
	return controllers.RenderTemplate(c, "friends", nil)
}

func searchUsers(c echo.Context) error {
	searchParam := c.FormValue("search")
	fmt.Println("()()()() \nquery received for: ", searchParam)

	query := `
    SELECT username, email
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
	}

	for rows.Next() {
		var (
			username sql.NullString
			email    sql.NullString
		)
		if err := rows.Scan(&username, &email); err != nil {
			fmt.Println("error scanning rows:", err)
			return err
		}

		fmt.Println(" \n\n username: ", username.String)
		users = append(users, struct {
			Username string
			Email    string
		}{
			Username: username.String,
			Email:    email.String,
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
		}
	}

	dataStruct := DataStruct{
		Data: users,
	}
	// pageData := PageData{
	// 	data: users,
	// }

	return controllers.RenderTemplate(c, "search-results", dataStruct)
}
