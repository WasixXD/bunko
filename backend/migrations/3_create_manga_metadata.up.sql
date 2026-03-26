CREATE TABLE IF NOT EXISTS manga_metadata (
    manga_id INTEGER PRIMARY KEY,
    localized_name TEXT,
    publication_status TEXT,
    summary TEXT,
    start_year INTEGER,
    start_month INTEGER,
    start_day INTEGER,
    author TEXT,
    art TEXT,
    web_link TEXT,
    metadata_updated_at TIMESTAMP,
    FOREIGN KEY (manga_id)
        REFERENCES mangas (manga_id)
        ON DELETE CASCADE
);

INSERT OR IGNORE INTO manga_metadata (
    manga_id,
    localized_name,
    publication_status,
    summary,
    start_year,
    start_month,
    start_day,
    author,
    web_link,
    metadata_updated_at
)
SELECT
    manga_id,
    localized_name,
    publication_status,
    summary,
    start_year,
    start_month,
    start_day,
    author,
    web_link,
    metadata_updated_at
FROM mangas
WHERE localized_name IS NOT NULL
   OR publication_status IS NOT NULL
   OR summary IS NOT NULL
   OR start_year IS NOT NULL
   OR start_month IS NOT NULL
   OR start_day IS NOT NULL
   OR author IS NOT NULL
   OR web_link IS NOT NULL
   OR metadata_updated_at IS NOT NULL;
