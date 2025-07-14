package main

import (
	"io"
	"os"
	"time"

	"goodoo/database"
	"goodoo/handlers"
	"goodoo/http"
	"goodoo/logging"
	"goodoo/models"
	"goodoo/templates"

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

	// Initialize database
	dbName := os.Getenv("GOODOO_DEFAULT_DB")
	if dbName == "" {
		dbName = "apexive-hackaton"
	}
	
	logger.Info("Setting up database: %s", dbName)
	if err := database.QuickSetup(dbName, &models.User{}); err != nil {
		logger.Critical("Failed to setup database: %v", err)
		panic(err)
	}

	// Create default admin user if not exists
	initDefaultUser(dbName, logger)

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
		DefaultDBName:     dbName,
		SessionCookieName: "goodoo_session",
		Logger:            logger,
	}

	e := echo.New()

	// Set up template renderer
	e.Renderer = templates.NewTemplateRenderer()

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
	public.GET("/login", handlers.LoginPageHandler)
	public.GET("/health", healthHandler.Health)
	public.POST("/auth/login", authHandler.Login)
	public.GET("/db/list", dbHandler.ListDatabases)

	// Protected routes (authentication required)
	protected := e.Group("")
	protected.Use(http.AuthenticationMiddleware(true))
	protected.GET("/health/detailed", healthHandler.DetailedHealth)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/logout", authHandler.Logout)
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
	
	// Dashboard routes
	handlers.RegisterDashboardRoutes(e, requestConfig)

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

func initDefaultUser(dbName string, logger *logging.Logger) {
	db, err := database.GetDatabase(dbName)
	if err != nil {
		logger.Error("Failed to get database for user initialization: %v", err)
		return
	}

	// Check if admin user exists
	var count int64
	db.Model(&models.User{}).Where("login = ?", "admin").Count(&count)
	
	if count == 0 {
		logger.Info("Creating default admin user")
		_, err := models.CreateUser(db, "admin", "Administrator", "admin@example.com", "admin")
		if err != nil {
			logger.Error("Failed to create default admin user: %v", err)
		} else {
			logger.Info("Default admin user created successfully (login: admin, password: admin)")
		}
	} else {
		logger.Info("Admin user already exists")
	}
}
