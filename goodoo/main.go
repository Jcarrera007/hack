package main

import (
	"io"
	"os"
	"time"

	"goodoo/handlers"
	"goodoo/http"
	"goodoo/logging"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize logging system
	if err := logging.InitLogger(); err != nil {
		panic(err)
	}

	logger := logging.GetLogger("goodoo.main")
	logger.Info("Starting Goodoo application")

	// Initialize session store
	sessionDir := os.Getenv("GOODOO_SESSION_DIR")
	if sessionDir == "" {
		sessionDir = "./sessions"
	}

	sessionStore, err := http.NewFilesystemSessionStore(sessionDir, true)
	if err != nil {
		logger.Critical("Failed to create session store: %v", err)
		panic(err)
	}

	// Create request configuration
	requestConfig := &http.RequestConfig{
		SessionStore:      sessionStore,
		DefaultDBName:     os.Getenv("GOODOO_DEFAULT_DB"),
		SessionCookieName: "goodoo_session",
		Logger:            logger,
	}

	e := echo.New()

	// Disable Echo's default logger since we have our own
	e.Logger.SetOutput(io.Discard)

	// Core middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Goodoo middleware
	e.Use(http.RequestMiddleware(requestConfig))
	e.Use(logging.PerformanceMiddleware())
	e.Use(http.SecurityMiddleware())
	e.Use(http.ErrorHandlingMiddleware())
	e.Use(http.RequestLoggingMiddleware())

	// Session cleanup (every hour)
	e.Use(http.SessionCleanupMiddleware(sessionStore, 1*time.Hour))

	// Static files
	e.Static("/static", "static")

	// Create handlers
	authHandler := handlers.NewAuthHandler(requestConfig)
	dbHandler := handlers.NewDatabaseHandler(requestConfig)
	healthHandler := handlers.NewHealthHandler(requestConfig)
	sessionHandler := handlers.NewSessionHandler(requestConfig)

	// Public routes (no authentication required)
	public := e.Group("")
	public.GET("/", handlers.IndexHandler)
	public.GET("/health", healthHandler.Health)
	public.POST("/auth/login", authHandler.Login)
	public.GET("/db/list", dbHandler.ListDatabases)

	// Protected routes (authentication required)
	protected := e.Group("")
	protected.Use(http.AuthenticationMiddleware(true))
	protected.GET("/health/detailed", healthHandler.DetailedHealth)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/session", authHandler.SessionInfo)
	protected.POST("/db/set", dbHandler.SetDatabase)
	protected.GET("/session", sessionHandler.GetSession)
	protected.POST("/session/clear", sessionHandler.ClearSession)
	protected.POST("/session/set", sessionHandler.SetSessionData)

	// Database-dependent routes
	withDB := e.Group("")
	withDB.Use(http.AuthenticationMiddleware(true))
	withDB.Use(http.DatabaseMiddleware(true))
	// Add database-dependent routes here

	// API routes
	handlers.RegisterAPIRoutes(e)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server on port %s", port)
	logger.Info("Session store: %s", sessionDir)
	logger.Info("Default database: %s", requestConfig.DefaultDBName)

	if err := e.Start(":" + port); err != nil {
		logger.Critical("Server failed to start: %v", err)
	}
}
