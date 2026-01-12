package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/urfave/cli/v3"
)

var MigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "Database migration commands",
	Commands: []*cli.Command{
		{
			Name:      "up",
			Usage:     "Run migrations up (optionally to a specific version)",
			ArgsUsage: "[version]",
			Action:    migrateUp,
		},
		{
			Name:      "down",
			Usage:     "Revert migrations down (optionally to a specific version)",
			ArgsUsage: "[version]",
			Action:    migrateDown,
		},
	},
}

func getMigrateArgs() []string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost:5432/hyperwhisper?sslmode=disable"
	}

	return []string{
		"-path", "migrations",
		"-database", dbURL,
	}
}

func runMigrate(args ...string) error {
	baseArgs := getMigrateArgs()
	allArgs := append(baseArgs, args...)

	cmd := exec.Command("migrate", allArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func migrateUp(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args()

	if args.Len() == 0 {
		// Run all migrations
		fmt.Println("Running all pending migrations...")
		return runMigrate("up")
	}

	// Run up to specific version
	versionStr := args.First()
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version number: %s", versionStr)
	}

	fmt.Printf("Running migrations up to version %d...\n", version)
	return runMigrate("goto", strconv.Itoa(version))
}

func migrateDown(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args()

	if args.Len() == 0 {
		// Revert all migrations
		fmt.Println("Reverting all migrations...")
		return runMigrate("down", "-all")
	}

	// Revert to specific version
	versionStr := args.First()
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version number: %s", versionStr)
	}

	fmt.Printf("Reverting migrations down to version %d...\n", version)
	return runMigrate("goto", strconv.Itoa(version))
}
