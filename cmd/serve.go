package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/urfave/cli/v3"
)

var ServeCommand = &cli.Command{
	Name:  "serve",
	Usage: "Start the server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "api-port",
			Value: "1323",
			Usage: "Port for the API server",
		},
		&cli.StringFlag{
			Name:  "api-host",
			Value: "0.0.0.0",
			Usage: "Host for the API server",
		},
		&cli.BoolFlag{
			Name:  "dev",
			Value: false,
			Usage: "Run in development mode (starts nuxt dev server)",
		},
	},
	Action: runServe,
}

func runServe(ctx context.Context, cmd *cli.Command) error {
	host := cmd.String("api-host")
	port := cmd.String("api-port")
	dev := cmd.Bool("dev")

	var nuxtCmd *exec.Cmd

	if dev {
		// Start nuxt dev server
		nuxtCmd = exec.Command("bun", "run", "dev")
		nuxtCmd.Dir = "web"
		nuxtCmd.Stdout = os.Stdout
		nuxtCmd.Stderr = os.Stderr

		if err := nuxtCmd.Start(); err != nil {
			return fmt.Errorf("failed to start nuxt dev server: %w", err)
		}

		fmt.Println("Started Nuxt dev server on port 3000")
	}

	// Setup Echo server
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// API routes group
	api := e.Group("/api/v1")
	setupAPIRoutes(api)

	// Proxy non-API requests to Nuxt
	nuxtURL, _ := url.Parse("http://localhost:3000")
	proxy := httputil.NewSingleHostReverseProxy(nuxtURL)

	e.Any("/*", func(c echo.Context) error {
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nShutting down...")

		if nuxtCmd != nil && nuxtCmd.Process != nil {
			nuxtCmd.Process.Signal(syscall.SIGTERM)
			nuxtCmd.Wait()
		}

		e.Shutdown(context.Background())
	}()

	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Starting API server on %s\n", addr)

	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func setupAPIRoutes(api *echo.Group) {
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}
