package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/urfave/cli/v3"

	"hyperwhisper/migrations"
)

var MigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "Database migration commands",
	Commands: []*cli.Command{
		{
			Name:      "up",
			Usage:     "Run migrations up (optionally specify number of steps)",
			ArgsUsage: "[steps]",
			Action:    migrateUp,
		},
		{
			Name:      "down",
			Usage:     "Revert migrations down (optionally specify number of steps)",
			ArgsUsage: "[steps]",
			Action:    migrateDown,
		},
		{
			Name:      "version",
			Usage:     "Print current migration version",
			Action:    migrateVersion,
		},
		{
			Name:      "goto",
			Usage:     "Migrate to a specific version",
			ArgsUsage: "<version>",
			Action:    migrateGoto,
		},
	},
}

func getDBURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost:5432/hyperwhisper?sslmode=disable"
	}
	return dbURL
}

func newMigrate() (*migrate.Migrate, error) {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}
	return migrate.NewWithSourceInstance("iofs", source, getDBURL())
}

func migrateUp(ctx context.Context, cmd *cli.Command) error {
	m, err := newMigrate()
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer m.Close()

	args := cmd.Args()

	if args.Len() == 0 {
		fmt.Println("Running all pending migrations...")
		err = m.Up()
	} else {
		steps, err := strconv.Atoi(args.First())
		if err != nil {
			return fmt.Errorf("invalid steps number: %s", args.First())
		}
		fmt.Printf("Running %d migration(s) up...\n", steps)
		err = m.Steps(steps)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("No migrations to apply.")
		return nil
	}

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations applied successfully.")
	return nil
}

func migrateDown(ctx context.Context, cmd *cli.Command) error {
	m, err := newMigrate()
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer m.Close()

	args := cmd.Args()

	if args.Len() == 0 {
		fmt.Println("Reverting all migrations...")
		err = m.Down()
	} else {
		steps, err := strconv.Atoi(args.First())
		if err != nil {
			return fmt.Errorf("invalid steps number: %s", args.First())
		}
		fmt.Printf("Reverting %d migration(s)...\n", steps)
		err = m.Steps(-steps)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("No migrations to revert.")
		return nil
	}

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations reverted successfully.")
	return nil
}

func migrateVersion(ctx context.Context, cmd *cli.Command) error {
	m, err := newMigrate()
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("No migrations have been applied yet.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	fmt.Printf("Current version: %d", version)
	if dirty {
		fmt.Print(" (dirty)")
	}
	fmt.Println()

	return nil
}

func migrateGoto(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("version argument is required")
	}

	version, err := strconv.ParseUint(args.First(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid version number: %s", args.First())
	}

	m, err := newMigrate()
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer m.Close()

	fmt.Printf("Migrating to version %d...\n", version)
	err = m.Migrate(uint(version))

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("Already at target version.")
		return nil
	}

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migration completed successfully.")
	return nil
}
