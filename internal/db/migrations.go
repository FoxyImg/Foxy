package db

import (
	"context"
	"embed"
	"foxy/internal/env"
	"github.com/jackc/pgx/v5"
	"log"
	"slices"
)

//go:embed up/*.sql
var migrationsFS embed.FS

var migrations = []string{
	"up.2024.05.21.sql",
	"up.2024.05.23.sql",
	"up.2024.05.27.sql",
	"up.2024.06.06.sql",
}

func RunMigrations() error {
	if env.FoxyEnvironment.ServerType != "primary" {
		log.Println("Not a primary server.  Migrations won't run.")
		return nil
	}

	if env.FoxyEnvironment.DatabaseUrl == nil {
		log.Println("No database url set.  Migrations won't run.")
		return nil
	}

	if len(migrations) == 0 {
		log.Println("No DB up migration files found.  Migrations won't run.")
		return nil
	}

	conn, err := NewConnection()
	if err != nil {
		log.Println("Error connecting to database.  Migrations won't run.", err)
		return nil
	}

	defer conn.Release()

	var previousMigrations []string
	res, err := conn.Query(context.Background(), "SELECT name FROM migrations")
	if err == nil {
		var name string
		_, _ = pgx.ForEachRow(res, []any{&name}, func() error {
			previousMigrations = append(previousMigrations, name)
			return nil
		})
		res.Close()
	}

	log.Println("Running migrations", migrations)

	for _, file := range migrations {
		if slices.Contains(previousMigrations, file) {
			log.Println("Skipping migration", file)
			continue
		}

		text, err := migrationsFS.ReadFile("up/" + file)
		if err != nil {
			log.Println("Read File Error:", err)
			return err
		}

		log.Println("Running migration", string(text))
		_, err = conn.Exec(context.Background(), string(text))
		if err != nil {
			log.Println("Exec Error:", err)
			return err
		}

		_, err = conn.Exec(context.Background(), "INSERT INTO migrations (name) VALUES ($1)", file)
		if err != nil {
			log.Println("Exec Error:", err)
			return err
		}
	}

	return nil

}
