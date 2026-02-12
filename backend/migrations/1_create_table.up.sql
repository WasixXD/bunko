
CREATE TABLE IF NOT EXISTS `mangas` (
    manga_id integer primary key,

    name text not null,
    slug text not null,
    status text not null check (status in ("downloading", "pending", "completed")),
    provider text not null,
    url text not null,
    cover_path text,
    manga_path text not null,

    localized_name text,
    publication_status text,
    summary text,
    start_year int,
    start_month int,
    start_day int,
    author text,
    web_link text,
    metadata_updated_at timestamp,

    created_at timestamp default (datetime('now'))
);

CREATE TABLE IF NOT EXISTS `chapter` (
    manga_id integer not null,
    chapter_id integer primary key,
    url text not null,
    name text not null,
    foreign key (manga_id)
        references mangas (manga_id)
);


CREATE TABLE IF NOT EXISTS `download_queue` (
    manga_id integer not null,
    name text not null,
    url text not null,
    status text not null default 'pending' check (status in ("downloading", "pending", "completed", "error")),
    provider text not null,
    path_to_download text not null,
    foreign key (manga_id)
        references mangas (manga_id)
);
