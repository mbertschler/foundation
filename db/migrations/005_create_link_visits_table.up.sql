CREATE TABLE IF NOT EXISTS link_visits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_link TEXT NOT NULL,
    user_id INTEGER,
    visited_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f','now')),
    FOREIGN KEY(short_link) REFERENCES links(short_link),
    FOREIGN KEY(user_id) REFERENCES users(id)
);
