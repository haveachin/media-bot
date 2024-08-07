CREATE TABLE IF NOT EXISTS requests (
    id INTEGER PRIMARY KEY,
    requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username TEXT NOT NULL,
    requested_url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS file_infos (
    id BLOB NOT NULL PRIMARY KEY,
    request_id INTEGER NOT NULL,
    size INTEGER NOT NULL,
    hash BLOB NOT NULL UNIQUE,
    url TEXT,
    FOREIGN KEY(request_id) REFERENCES requests(id)
);