// Package main is a tiny CLI wrapper around golang-migrate to run SQL migrations.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/instaagrammeta/somon-crm/backend/internal/config"
)

func main() {
	dir := flag.String("dir", "migrations", "Migrations directory")
	direction := flag.String("direction", "up", "up | down | force | version")
	steps := flag.Int("steps", 0, "Number of steps (0 = all)")
	versionForce := flag.Int("version", 0, "Version to force (used with -direction=force)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode)

	m, err := migrate.New("file://"+*dir, dsn)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	defer m.Close()

	switch *direction {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
	case "force":
		err = m.Force(*versionForce)
	case "version":
		v, dirty, e := m.Version()
		if e != nil {
			log.Fatal(e)
		}
		log.Printf("version=%d dirty=%v", v, dirty)
		return
	default:
		log.Fatalf("unknown direction: %s", *direction)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate %s: %v", *direction, err)
	}
	log.Printf("migrate %s: ok", *direction)
}
