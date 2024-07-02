WITH games_by_user_id AS (
  SELECT *
  FROM user_game_membership
  WHERE :UserId = user_game_membership.user_id
)
