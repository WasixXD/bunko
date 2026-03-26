package downloader

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"bunko/backend/structs"
)

func (d *Downloader) TurnIntoCbz(source string) error {
	target := fmt.Sprintf("%s.cbz", source)

	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()
	defer os.RemoveAll(source)

	err = filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func (d *Downloader) CreateComicInfo(mangaID int, chapterName, chapterPath string) error {
	comic := structs.ComicInfo{}
	const query = `
		SELECT 
			COALESCE(m.name, '') AS series,
			COALESCE(md.localized_name, '') AS localized_series,
			COALESCE(md.web_link, '') AS web,
			CASE 
				WHEN md.publication_status = 'RELEASING' THEN 1
				ELSE 0
			END AS publication_status,
			COALESCE(md.summary, '') AS summary,
			COALESCE(md.author, '') AS writer,
			COALESCE(md.start_year, 0) AS year,
			COALESCE(md.start_month, 0) AS month,
			COALESCE(md.start_day, 0) AS day
		FROM mangas m
		LEFT JOIN manga_metadata md ON md.manga_id = m.manga_id
		WHERE m.manga_id = ?
	`

	if err := d.Database.Get(&comic, query, mangaID); err != nil {
		return err
	}

	comic.Title = chapterName

	info, err := xml.MarshalIndent(comic, " ", "  ")
	if err != nil {
		return err
	}

	comicInfoPath := filepath.Join(chapterPath, "comicinfo.xml")
	return os.WriteFile(comicInfoPath, info, 0755)
}
