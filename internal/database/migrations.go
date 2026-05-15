package database

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL string) error {
	log.Println("Starting database migrations...")

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Printf("Migration instance creation failed: %v", err)
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database schema is already up to date")
			return nil
		}
		log.Printf("Migration execution failed: %v", err)
		return err
	}

	log.Println("Database migrations applied successfully")
	return nil
}
