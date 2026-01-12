package main

import (
	"context"
	"fmt"
	"os"

	"hyperwhisper/cmd"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "hyperwhisper",
		Usage: "HyperWhisper server CLI",
		Commands: []*cli.Command{
			cmd.ServeCommand,
			cmd.MigrateCommand,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
