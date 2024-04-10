CREATE TABLE variable (
    id              INTEGER         NOT NULL    PRIMARY KEY,
    key             TEXT            NOT NULL    UNIQUE,
    value           INTEGER
);

CREATE TABLE account (
	id            INTEGER          NOT NULL     PRIMARY KEY,
	username      TEXT UNIQUE      NOT NULL,
	email         TEXT UNIQUE      NOT NULL,
	hash          TEXT             NOT NULL,
	token         TEXT UNIQUE,
	token_issued  INTEGER,
	created       INTEGER          NOT NULL,
	verified      BOOLEAN          NOT NULL,

	class         INTEGER          NOT NULL
);

