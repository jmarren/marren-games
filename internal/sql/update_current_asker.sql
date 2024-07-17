UPDATE current_askers
SET user_id = (
  SELECT (
    CASE
        WHEN (
          SELECT COUNT(*)
          FROM user_game_membership
          WHERE game_id = current_askers.game_id
            AND user_game_membership.user_id > current_askers.user_id
        ) > 0 THEN (
          SELECT user_id
          FROM user_game_membership
          WHERE current_askers.game_id = user_game_membership.game_id
            AND user_game_membership.user_id > current_askers.user_id
          ORDER BY user_game_membership.user_id
          LIMIT 1
          )
        ELSE (
          SELECT user_id
          FROM user_game_membership
          WHERE user_game_membership.game_id = current_askers.game_id
          ORDER BY user_game_membership.user_id
          LIMIT 1
        )
    END
  )
  FROM user_game_membership
);
