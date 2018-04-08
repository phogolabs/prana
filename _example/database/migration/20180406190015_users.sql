-- Auto-generated at Fri Apr  6 19:00:15 CEST 2018
-- Please do not change the name attributes

-- name: up
CREATE TABLE users (
  id INT PRIMARY KEY NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT
);

-- name: down
DROP TABLE IF EXISTS users;

