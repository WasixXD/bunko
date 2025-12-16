
CREATE TABLE IF NOT EXISTS `mangas` (
    manga_id integer primary key,
    anilist_id integer,
    name text not null,
    slug text not null,
    status text not null check (status in ("downloading", "pending", "completed")),
    provider text not null,
    created_at timestamp default (datetime('now'))
);

CREATE TABLE IF NOT EXISTS `chapter` (
    manga_id integer not null,
    chapter_id integer primary key,
    chapter_n integer not null,
    url text not null,
    foreign key (manga_id)
        references mangas (manga_id)
);


CREATE TABLE IF NOT EXISTS `download_queue` (
    chapter_id
    url text,
    status text not null default 'pending',
    foreign key (chapter_id)
        references chapter (chapter_id)
);
