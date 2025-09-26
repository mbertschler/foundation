CREATE TABLE IF NOT EXISTS links (
    short_link TEXT PRIMARY KEY,
    full_url TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f','now')),
    FOREIGN KEY(user_id) REFERENCES users(id)
);
