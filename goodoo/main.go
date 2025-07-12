package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"goodoo/handlers"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Static files
	e.Static("/static", "static")

	// Routes
	e.GET("/", handlers.IndexHandler)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"app":    "goodoo",
		})
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}