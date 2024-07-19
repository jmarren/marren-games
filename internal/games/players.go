package games

import (
	"database/sql"
	"fmt"

	"github.com/jmarren/marren-games/internal/db"
)

func AddPlayer(tx *sql.Tx, gameId int, userId int) error {
	gameIdArg := sql.Named("game_id", gameId)
	userIdArg := sql.Named("user_id", userId)

	query := `
    INSERT INTO user_game_membership (user_id, game_id)
    VALUES (:user_id, :game_id);
  `

	_, err := db.Sqlite.Exec(query, userIdArg, gameIdArg)
	if err != nil {
		return fmt.Errorf("error: AddPlayer: %v:", err)
	}

	return nil
}
