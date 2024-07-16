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
  name TEXT NOT NULL,
  creator_id INTEGER NOT NULL,
  FOREIGN KEY (creator_id) REFERENCES users(id)
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
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  game_id  INTEGER NOT NULL,
  answer_text TEXT NOT NULL,
  date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  question_id INTEGER NOT NULL,
  answerer_id INTEGER NOT NULL,
  FOREIGN KEY (game_id) REFERENCES games(id),
  FOREIGN KEY (question_id) REFERENCES questions(id),
  FOREIGN KEY (answerer_id) REFERENCES users(id)
);



CREATE TABLE IF NOT EXISTS votes (
  game_id  INTEGER NOT NULL,
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  voter_id INTEGER NOT NULL,
  question_id INTEGER NOT NULL,
  answer_id INTEGER NOT NULL,
  FOREIGN KEY (game_id) REFERENCES games(id),
  FOREIGN KEY (voter_id) REFERENCES users(id),
  FOREIGN KEY (question_id) REFERENCES questions(id),
  FOREIGN KEY (answer_id) REFERENCES answers(id)
);


CREATE VIEW IF NOT EXISTS todays_question_id AS
    SELECT questions.id
    FROM questions
    WHERE DATE(questions.date_created) = DATE('now');


 CREATE VIEW IF NOT EXISTS todays_answers AS
    SELECT  answers.answer_text,
            answers.id AS answer_id,
            users.id AS answerer_id,
            users.username AS answerer_username,
            questions.id AS question_id
    FROM answers
    JOIN users
      ON users.id = answer_id
    JOIN questions
      ON question_id = answers.question_id
    WHERE DATE(answers.date_created) = DATE('now')
    AND DATE(questions.date_created) = DATE('now');



CREATE VIEW IF NOT EXISTS todays_question AS
    SELECT questions.question_text, users.username AS asker_username, questions.id AS question_id
    FROM questions
    JOIN users
      ON users.id = questions.asker_id
    WHERE DATE(questions.date_created) = DATE('now')
    LIMIT  1;


CREATE VIEW IF NOT EXISTS todays_game AS
    SELECT todays_question.question_text,
           todays_question.asker_username,
            GROUP_CONCAT('{"answerer_username":' ||  todays_answers.answerer_username ||
              '", "answer_text":"' ||  todays_answers.answer_text ||  '}') AS answers
          FROM todays_question
          JOIN todays_answers
            ON todays_answers.question_id = todays_question.question_id;



  CREATE VIEW IF NOT EXISTS answers_by_votes AS
    SELECT
      answers.answer_text,
      answers.id,
      users.username AS answerer_username,
      COUNT(votes.id) as total_votes
    FROM answers
    JOIN users
      ON users.id = answers.answerer_id
    JOIN votes
      ON votes.answer_id = answers.id
    WHERE  DATE(answers.date_created) = DATE('now')
    GROUP BY answers.id
    ORDER BY total_votes;

