CREATE TABLE feed (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    filepath TEXT NOT NULL,
    channels TEXT NOT NULL,
    datetime TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX feed_by_datetime ON feed (datetime);
