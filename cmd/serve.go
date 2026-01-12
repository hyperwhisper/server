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

	"hyperwhisper/internal/auth"
	"hyperwhisper/internal/db"
	"hyperwhisper/internal/handlers"

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

	// Connect to database
	if err := db.Connect(); err != nil {
		fmt.Printf("Warning: Could not connect to database: %v\n", err)
	} else {
		defer db.Close()
	}

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

	api.GET("/ht", healthCheck)

	// Auth routes (public)
	authHandler := handlers.NewAuthHandler(db.DB)
	api.POST("/signup", authHandler.SignUp)
	api.POST("/signin", authHandler.SignIn)
	api.POST("/token_refresh", authHandler.TokenRefresh)
	api.POST("/signout", authHandler.SignOut)

	// Protected routes
	protected := api.Group("")
	protected.Use(auth.JWTMiddleware())
	protected.GET("/me", authHandler.Me)

	// Admin routes (protected + admin only)
	admin := api.Group("/admin")
	admin.Use(auth.JWTMiddleware())
	admin.Use(auth.AdminMiddleware())

	adminHandler := handlers.NewAdminHandler(db.DB)

	// User management
	admin.GET("/users", adminHandler.ListUsers)
	admin.POST("/users", adminHandler.CreateUser)
	admin.DELETE("/users/:id", adminHandler.DeleteUser)

	// Token management
	admin.GET("/tokens", adminHandler.ListRefreshTokens)
	admin.POST("/tokens/revoke", adminHandler.RevokeToken)
	admin.POST("/tokens/revoke-user/:id", adminHandler.RevokeUserRefreshTokens)
	admin.POST("/tokens/cleanup", adminHandler.CleanupTokens)
}

type HealthCheckResponse struct {
	All bool `json:"all"`
	DB  bool `json:"db"`
	API bool `json:"api"`
}

func healthCheck(c echo.Context) error {
	response := HealthCheckResponse{
		API: true,
		DB:  false,
		All: false,
	}

	if err := db.Ping(); err == nil {
		response.DB = true
	}

	response.All = response.API && response.DB

	status := http.StatusOK
	if !response.All {
		status = http.StatusServiceUnavailable
	}

	return c.JSON(status, response)
}
