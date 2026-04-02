package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		path = flag.String("path", "file://migrations", "migration source path")
	)

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("usage: go run ./cmd/migrate [up|down|version]")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New(*path, databaseURL)
	if err != nil {
		log.Fatalf("create migrate instance: %v", err)
	}

	command := flag.Arg(0)

	switch command {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("apply up migrations: %v", err)
		}
		fmt.Println("migrations are up to date")
	case "down":
		err = m.Steps(-1)
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("apply down migration: %v", err)
		}
		fmt.Println("rolled back one migration")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				fmt.Println("no migrations applied")
				return
			}

			log.Fatalf("read migration version: %v", err)
		}

		fmt.Printf("version=%d dirty=%t\n", version, dirty)
	default:
		log.Fatalf("unknown command %q", command)
	}
}
