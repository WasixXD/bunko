PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS mangas (
    manga_id INTEGER PRIMARY KEY,

    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('downloading', 'pending', 'completed')),
    provider TEXT NOT NULL,
    url TEXT NOT NULL,
    cover_path TEXT,
    manga_path TEXT NOT NULL,

    localized_name TEXT,
    publication_status TEXT,
    summary TEXT,
    start_year INTEGER,
    start_month INTEGER,
    start_day INTEGER,
    author TEXT,
    web_link TEXT,
    metadata_updated_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS chapter (
    chapter_id INTEGER PRIMARY KEY,
    manga_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    name TEXT NOT NULL,

    FOREIGN KEY (manga_id)
        REFERENCES mangas (manga_id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cron (
    manga_id INTEGER NOT NULL,
    rule TEXT,
    last_updated_at TIMESTAMP DEFAULT (datetime('now')),
    FOREIGN KEY (manga_id)
        REFERENCES mangas (manga_id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS download_queue (
    manga_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('downloading', 'pending', 'completed', 'error')),
    provider TEXT NOT NULL,
    path_to_download TEXT NOT NULL,

    FOREIGN KEY (manga_id)
        REFERENCES mangas (manga_id)
        ON DELETE CASCADE
);