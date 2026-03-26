## Bunko

Bunko is a self hosted application that can download mangas and manage metadata

# <img width="1895" height="979" alt="home_bunko" src="https://github.com/user-attachments/assets/8b960db4-9460-492d-8a17-3871aba682ba" />

|                   Adding a Manga                |             Status page                    |
| :---------------------------------------------: | :----------------------------------------: |
| <img width="1916" height="985" alt="add-manga-bunko" src="https://github.com/user-attachments/assets/50c42691-2812-4c35-a5c4-33f1ea8ba1ff" /> | <img width="1876" height="822" alt="death_nopte" src="https://github.com/user-attachments/assets/a9010d66-091e-48ac-b55a-9129f709115a" /> |


## Deployment

```docker
services:
  bunko:
    container_name: bunko
    image: lucaswasilewski/bunko:${BUNKO_VERSION:-latest}
    pull_policy: always
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      TZ: ${TZ:-UTC}
      BUNKO_DATABASE: /data/bunko.db
      BUNKO_MANGA_PATH: /library 
    volumes:
      - ./data:/data
      - ./library:/library
```

## Development

- go +v1.24
- node +24
- npm +11.26


## Notes on AI

With the exception of the frontend the entirety of this project was made by me. AI tooling was used to lint, organize and format the project.

## Acknowledgments

This project was heavily inspired by tools as:
- [Kaizoku](https://github.com/oae/kaizoku)
- [Mangal](https://github.com/metafates/mangal)
