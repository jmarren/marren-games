
PRAGMA foreign_keys = ON;


CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS friendships (
  user_1_id INTEGER NOT NULL,
  user_2_id INTEGER NOT NULL,
  FOREIGN KEY (user_1_id) REFERENCES users(id),
  FOREIGN KEY (user_2_id) REFERENCES users(id),
  PRIMARY KEY (user_1_id, user_2_id)
);

CREATE TABLE IF NOT EXISTS friend_requests (
  from_user_id INTEGER NOT NULL,
  to_user_id INTEGER NOT NULL,
  FOREIGN KEY (from_user_id) REFERENCES users(id),
  FOREIGN KEY (to_user_id) REFERENCES users(id),
  PRIMARY KEY (from_user_id, to_user_id)
);

CREATE TABLE IF NOT EXISTS games (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  name TEXT NOT NULL UNIQUE,
  creator_id INTEGER NOT NULL,
  FOREIGN KEY (creator_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS current_askers (
  user_id INTEGER NOT NULL,
  game_id INTEGER NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (game_id) REFERENCES games(id),
  PRIMARY KEY (game_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_game_membership (
  user_id INTEGER NOT NULL,
  game_id INTEGER NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (game_id) REFERENCES games(id),
  PRIMARY KEY (user_id, game_id)
);



CREATE TABLE IF NOT EXISTS user_game_invites (
  user_id INTEGER NOT NULL,
  game_id INTEGER NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (game_id) REFERENCES games(id),
  PRIMARY KEY (user_id, game_id)
);

CREATE TABLE IF NOT EXISTS questions (
  game_id  INTEGER NOT NULL,
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  question_text TEXT NOT NULL,
  option_1 TEXT NOT NULL,
  option_2 TEXT NOT NULL,
  option_3 TEXT NOT NULL,
  option_4 TEXT NOT NULL,
  date_created  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  asker_id INTEGER NOT NULL,
  FOREIGN KEY (asker_id) REFERENCES users(id),
  FOREIGN KEY (game_id) REFERENCES games(id)
);

CREATE TABLE IF NOT EXISTS answers (
  game_id  INTEGER NOT NULL,
  option_chosen INTEGER CHECK(option_chosen IN (1,2,3,4)),
  question_id INTEGER NOT NULL,
  answerer_id INTEGER NOT NULL,
  FOREIGN KEY (game_id) REFERENCES games(id),
  FOREIGN KEY (question_id) REFERENCES questions(id),
  FOREIGN KEY (answerer_id) REFERENCES users(id),
  PRIMARY KEY (game_id, question_id, answerer_id)
);

CREATE TABLE IF NOT EXISTS scores (
  user_id INTEGER NOT NULL,
  game_id INTEGER NOT NULL,
  score   INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (game_id) REFERENCES games(id),
  PRIMARY KEY (user_id, game_id)
);

CREATE TRIGGER IF NOT EXISTS add_new_user_to_all_users_game
AFTER INSERT ON users
BEGIN
  INSERT OR IGNORE INTO user_game_membership (user_id, game_id)
  SELECT NEW.id, id
  FROM games
  WHERE name = "All Users";
END;

CREATE TRIGGER IF NOT EXISTS new_game_insert
AFTER INSERT ON games
BEGIN
  -- add creator as a member of the game
  INSERT INTO user_game_membership (user_id, game_id)
  VALUES (NEW.creator_id, NEW.id);
  
  -- make the creator the current asker
  INSERT INTO current_askers (user_id, game_id)
  VALUES (NEW.creator_id, NEW.id);
END;


CREATE TRIGGER IF NOT EXISTS  insert_new_member_into_scores
AFTER INSERT ON user_game_membership
BEGIN
  -- add user to scores with a score of 0 (default)
  INSERT INTO scores (user_id, game_id)
  VALUES (NEW.user_id, NEW.game_id);
END;


CREATE VIEW IF NOT EXISTS todays_question_id AS
    SELECT questions.id
    FROM questions
    WHERE DATE(questions.date_created) = DATE('now');



CREATE VIEW IF NOT EXISTS todays_question AS
    SELECT questions.question_text, users.username AS asker_username, questions.id AS question_id
    FROM questions
    JOIN users
      ON users.id = questions.asker_id
    WHERE DATE(questions.date_created) = DATE('now')
    LIMIT  1;



