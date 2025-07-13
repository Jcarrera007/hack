package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func IndexHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Welcome to Goodoo Framework",
		"framework": "Go + Echo + Odoo-inspired patterns",
		"features": []string{
			"API decorators",
			"Field system with validation",
			"Session management",
			"Logging with performance tracking",
			"Database integration",
		},
		"endpoints": map[string]string{
			"health":     "/health",
			"api_docs":   "/api/models",
			"auth":       "/auth",
			"session":    "/session",
			"databases":  "/db",
		},
	})
}