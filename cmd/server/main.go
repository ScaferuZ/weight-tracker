package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weight-tracker/internal/config"
	"weight-tracker/internal/handlers"
	"weight-tracker/internal/middleware"
	"weight-tracker/internal/models"
)

type application struct {
	config *config.Config
	db     *sql.DB
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	database, err := config.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	app := &application{
		config: cfg,
		db:     database.GetDB(),
	}

	// Initialize handlers
	pageHandler := handlers.NewPageHandler()
	authHandler := handlers.NewAuthHandler(app.db)
	weightHandler := handlers.NewWeightHandler(app.db)
	chartHandler := handlers.NewChartHandler(app.db)
	healthHandler := handlers.NewHealthHandler(app.db)

	// Setup middleware
	authMiddleware := middleware.AuthMiddleware(models.NewUserRepository(app.db))

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Public routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			pageHandler.Home(w, r)
			return
		}
		pageHandler.NotFound(w, r)
	})

	mux.HandleFunc("/login", authHandler.ShowLogin)
	mux.HandleFunc("/register", authHandler.ShowRegister)
	mux.HandleFunc("/logout", authHandler.Logout)
mux.HandleFunc("/health", healthHandler.Health)

	// Protected routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/weights", weightHandler.ShowWeights)
	protectedMux.HandleFunc("/weights/", weightHandler.CreateWeight)
	protectedMux.HandleFunc("/api/chart/weight-data", chartHandler.GetWeightChartData)
	protectedMux.HandleFunc("/api/chart/weight-stats", chartHandler.GetWeightStats)

	// Apply auth middleware to all routes to set context values
	// This ensures that all requests have proper context values set
	allRoutesMux := authMiddleware(mux)

	// Create a handler that routes between protected and public routes
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Protected routes - require authentication
		if r.URL.Path == "/weights" ||
			(r.URL.Path == "/weights/" && r.Method == http.MethodPost) ||
			r.URL.Path == "/api/chart/weight-data" ||
			r.URL.Path == "/api/chart/weight-stats" {

			// Use RequireAuth middleware to protect these routes
			// RequireAuth checks the context values already set by authMiddleware
			protectedHandler := middleware.RequireAuth(protectedMux)
			protectedHandler.ServeHTTP(w, r)
			return
		}

		// Public routes - serve through the auth middleware (which sets context but doesn't require login)
		allRoutesMux.ServeHTTP(w, r)
	})

	// Apply global middleware
	var handler http.Handler = finalHandler
	handler = middleware.Logging(middleware.SecurityHeaders(handler))

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	log.Printf("Database: %s", cfg.DatabasePath)
	log.Printf("Environment: %s", cfg.Env)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}