package downloader

import (
	"archive/zip"
	"bunko/backend/db"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
	comic, err := db.GetComicInfo(d.Database, mangaID)
	if err != nil {
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
