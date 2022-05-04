DROP TABLE IF EXISTS feeds;

CREATE TABLE feeds (
  id VARCHAR(255) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description VARCHAR(255) NOT NULL,
  create_at TIMESTAMP NOT NULL DEFAULT NOW()
);
