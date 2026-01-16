package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"hyperwhisper/internal/auth"
	"hyperwhisper/internal/db"
	"hyperwhisper/internal/handlers"
	"hyperwhisper/web"

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

	if dev {
		// Proxy non-API requests to Nuxt dev server
		nuxtURL, _ := url.Parse("http://localhost:3000")
		proxy := httputil.NewSingleHostReverseProxy(nuxtURL)

		e.Any("/*", func(c echo.Context) error {
			proxy.ServeHTTP(c.Response(), c.Request())
			return nil
		})
	} else {
		// Serve embedded static files in production
		distFS, err := fs.Sub(web.DistFS, "dist")
		if err != nil {
			return fmt.Errorf("failed to get embedded dist folder: %w", err)
		}

		fileServer := http.FileServer(http.FS(distFS))

		e.Any("/*", func(c echo.Context) error {
			path := c.Request().URL.Path

			// Try to serve the exact file first
			if f, err := distFS.Open(strings.TrimPrefix(path, "/")); err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Response(), c.Request())
				return nil
			}

			// For SPA routing, serve index.html for non-file requests
			c.Request().URL.Path = "/"
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		})
	}

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

	// Trial handler for trial API keys
	trialHandler := handlers.NewTrialHandler(db.DB)

	// Deepgram routes
	deepgramHandler := handlers.NewDeepgramHandler(db.DB)

	// WebSocket endpoint (API key auth, not JWT)
	// This handler supports both regular API keys (hw_live_) and trial keys (hw_trial_)
	// Trial keys are automatically routed to the trial handler
	api.GET("/deepgram/listen", func(c echo.Context) error {
		// Inject trial handler into context for trial key routing
		c.Set("trial_handler", trialHandler)
		return deepgramHandler.DeepgramProxy(c)
	})

	// Dashboard WebSocket endpoint (JWT auth via cookie, no API key needed)
	// This endpoint has a 5-minute session limit and doesn't log to transcription_logs
	api.GET("/deepgram/dashboard/listen", deepgramHandler.DeepgramProxyDashboard, auth.JWTMiddleware())

	// API key management (JWT auth required)
	deepgram := api.Group("/deepgram")
	deepgram.Use(auth.JWTMiddleware())
	deepgram.POST("/keys", deepgramHandler.GenerateAPIKey)
	deepgram.GET("/keys", deepgramHandler.ListAPIKeys)
	deepgram.DELETE("/keys/:id", deepgramHandler.RevokeAPIKey)
	deepgram.GET("/usage", deepgramHandler.GetUsageSummary)
	deepgram.GET("/logs", deepgramHandler.ListTranscriptionLogs)

	// Trial routes (public, no JWT required)
	trial := api.Group("/trial")
	trial.POST("/provision", trialHandler.ProvisionTrialKey)
	trial.GET("/usage", trialHandler.GetTrialUsage)
	trial.GET("/status", trialHandler.GetTrialStatus)

	// Admin Deepgram routes
	admin.GET("/deepgram/logs", adminHandler.ListAllTranscriptionLogs)
	admin.GET("/deepgram/keys", adminHandler.ListAllAPIKeys)
	admin.GET("/deepgram/usage", adminHandler.GetSystemUsageSummary)

	// Admin Trial routes
	admin.GET("/trial/keys", adminHandler.ListTrialAPIKeys)
	admin.GET("/trial/usage", adminHandler.GetTrialUsageSummary)
	admin.GET("/trial/limits", adminHandler.GetTrialLimits)
	admin.PUT("/trial/limits", adminHandler.UpdateTrialLimits)
	admin.DELETE("/trial/keys/:id", adminHandler.RevokeTrialKey)
	admin.POST("/trial/cleanup", adminHandler.CleanupExpiredTrialKeys)
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
