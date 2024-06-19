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
  answer_id INTEGER NOT NULL,
  FOREIGN KEY (voter_id) REFERENCES users(id),
  FOREIGN KEY (question_id) REFERENCES questions(id),
  FOREIGN KEY (answer_id) REFERENCES answers(id)
);


CREATE VIEW IF NOT EXISTS todays_question_id AS
    SELECT questions.id
    FROM questions
    WHERE DATE(questions.date_created) = DATE('now');


CREATE VIEW IF NOT EXISTS todays_answers AS
    SELECT answers.answer_text, users.id AS answerer_id, users.username AS answerer_username, questions.id AS question_id
    FROM answers
    JOIN users
      ON users.id = answers.answerer_id
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
