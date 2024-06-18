PRAGMA foreign_keys = ON;



CREATE  TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS questions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  question_text TEXT NOT NULL,
  date_created  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  asker_id INTEGER NOT NULL,
  FOREIGN KEY (asker_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS answers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  answer_text TEXT NOT NULL,
  date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  question_id INTEGER NOT NULL,
  answerer_id INTEGER NOT NULL,
  FOREIGN KEY (question_id) REFERENCES questions(id),
  FOREIGN KEY (answerer_id) REFERENCES users(id)
);


CREATE TABLE IF NOT EXISTS votes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  voter_id INTEGER NOT NULL,
  question_id INTEGER NOT NULL,
  FOREIGN KEY (voter_id) REFERENCES users(id),
  FOREIGN KEY (question_id) REFERENCES questions(id)
);


CREATE VIEW IF NOT EXISTS todays_question_id AS
    SELECT questions.id
    FROM questions
    WHERE DATE(questions.date_created) = DATE('now');



