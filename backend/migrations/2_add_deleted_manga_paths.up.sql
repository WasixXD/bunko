CREATE TABLE IF NOT EXISTS deleted_manga_paths (
    path TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT (datetime('now'))
);
