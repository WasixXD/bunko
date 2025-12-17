package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func InitDb(dbPath string) (*sql.DB, error) {
	if err := checkOrCreate(dbPath); err != nil {
		return nil, fmt.Errorf("could not create sqlite3 file: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open sqlite3 file: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("could not execute migrations: %w", err)
	}

	return db, nil
}

func checkOrCreate(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func runMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %w", err)
	}

	migrationsPath, _ := filepath.Abs("./migrations")

	m, _ := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3",
		driver,
	)

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("could not migrate: %w", err)
	}

	log.Info("Migrations applied!")
	return nil
}
